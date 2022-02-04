package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/quay/zlog"
	"github.com/urfave/cli/v2"
)

var ReportsCmd = &cli.Command{
	Name:        "report",
	Description: "request reports for named containers",
	Usage:       "clair-load-test report",
	Action:      reportAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "host",
			Usage:   "--host localhost:6060/",
			Value:   "http://localhost:6060/",
			EnvVars: []string{"CLAIR_API"},
		},
		&cli.StringFlag{
			Name:    "containers",
			Usage:   "--containers ubuntu:latest,mysql:latest",
			Value:   "",
			EnvVars: []string{"CONTAINERS"},
		},
		&cli.StringFlag{
			Name:    "psk",
			Usage:   "--psk secretkey",
			Value:   "",
			EnvVars: []string{"PSK"},
		},
		&cli.BoolFlag{
			Name:    "delete",
			Usage:   "--delete",
			Value:   false,
			EnvVars: []string{"DELETE"},
		},
		&cli.DurationFlag{
			Name:    "timeout",
			Usage:   "--timeout 1m",
			Value:   time.Minute * 1,
			EnvVars: []string{"TIMEOUT"},
		},
		&cli.Float64Flag{
			Name:    "rate",
			Usage:   "--rate 1",
			Value:   1,
			EnvVars: []string{"RATE"},
		},
	},
}

type IndexReportReponse struct {
	Hash string `json:"manifest_hash"`
}

type reporter struct {
	host  string
	psk   string
	stats *Stats
	cl    *http.Client
}

func NewReporter(host, psk string) *reporter {
	return &reporter{
		host:  host,
		psk:   psk,
		stats: NewStats(),
		cl:    &http.Client{Timeout: time.Minute * 1},
	}
}

func reportAction(c *cli.Context) error {
	ctx := c.Context
	containersArg := c.String("containers")
	containers := strings.Split(containersArg, ",")
	psk := c.String("psk")
	host := c.String("host")
	delete := c.Bool("delete")
	timeout := c.Duration("timeout")
	perSecond := c.Float64("rate")

	reporter := NewReporter(host, psk)

	g, ctx := errgroup.WithContext(ctx)
	i := 0
	timer := time.NewTimer(timeout)
	ticker := time.NewTicker(time.Duration(1000/perSecond) * time.Millisecond)
loop:
	for {
		select {
		case <-timer.C:
			break loop
		case <-ticker.C:
			cc := containers[i]
			g.Go(func() error {
				err := reporter.reportForContainer(ctx, cc, delete)
				if err != nil {
					zlog.Error(ctx).Str("container", cc).Msg(err.Error())
					return nil
				}
				zlog.Debug(ctx).Str("container", cc).Msg("completed")
				return nil
			})
			i++
			if i+1 > len(containers) {
				i = 0
			}
		}
	}
	err := g.Wait()
	if err != nil {
		return err
	}

	stats := reporter.stats.GetStats()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	err = enc.Encode(stats)
	if err != nil {
		return err
	}
	return nil
}

func (r *reporter) reportForContainer(ctx context.Context, container string, delete bool) error {
	// Call clairctl for the manifest
	manifest, err := getManifest(ctx, container)
	if err != nil {
		return fmt.Errorf("could not generate manifest: %w", err)
	}
	// Get a token
	logout.Debug().Str("container", container).Msg("got manifest")
	token, err := createToken(r.psk)
	if err != nil {
		zlog.Debug(ctx).Str("PSK", r.psk).Msg("creating token")
		return fmt.Errorf("could not create token: %w", err)
	}
	// Send manifest as body to index_report
	hash, err := r.createIndexReport(ctx, manifest, token)
	if err != nil {
		return fmt.Errorf("could not create index report: %w", err)
	}
	// Get a token
	// Request vuln report
	err = r.getVulnerabilityReport(ctx, hash, token)
	if err != nil {
		return fmt.Errorf("could not get vulnerability report: %w", err)
	}
	// Delete index_report
	if delete {
		err = r.deleteIndexReports(ctx, hash, token)
		if err != nil {
			return fmt.Errorf("could not delete index report: %w", err)
		}
	}
	return nil
}

func getManifest(ctx context.Context, container string) ([]byte, error) {
	cmd := exec.Command("clairctl", "manifest", container)
	zlog.Debug(ctx).Str("container", cmd.String()).Msg("getting manifest")
	return cmd.Output()
}

func (r *reporter) createIndexReport(ctx context.Context, body []byte, token string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		r.host+"/indexer/api/v1/index_report",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// Start clock
	t := time.Now()
	resp, err := r.cl.Do(req)
	// end clock and report
	diff := time.Now().Sub(t)
	r.stats.IncrTotalIndexReportRequestLatencyMilliseconds(diff.Milliseconds())
	r.stats.IncrTotalIndexReportRequests(int64(1))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		r.stats.IncrNon2XXIndexReportResponses(int64(1))
		return "", fmt.Errorf("non 201 response from indexer %d", resp.StatusCode)
	}
	// decode response
	var irr = &IndexReportReponse{}
	err = json.NewDecoder(resp.Body).Decode(&irr)
	if err != nil {
		return "", err
	}

	return irr.Hash, nil
}

func (r *reporter) getVulnerabilityReport(ctx context.Context, hash string, token string) error {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet,
		r.host+"/matcher/api/v1/vulnerability_report/"+hash,
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	// Start clock
	t := time.Now()
	resp, err := r.cl.Do(req)
	// end clock and report
	diff := time.Now().Sub(t)
	r.stats.IncrTotalVulnerabilityReportRequestLatencyMilliseconds(diff.Milliseconds())
	r.stats.IncrTotalVulnerabilityReportRequests(int64(1))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		r.stats.IncrNon2XXVulnerabilityReportResponses(int64(1))
		return fmt.Errorf("non 200 response from matcher %d", resp.StatusCode)
	}
	return nil
}

func (r *reporter) deleteIndexReports(ctx context.Context, hash string, token string) error {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodDelete,
		r.host+"/indexer/api/v1/index_report/"+hash,
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	zlog.Debug(ctx).Str("hash", hash).Msg("deleting index report")
	resp, err := r.cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("non 204 response from indexer while deleting %d", resp.StatusCode)
	}
	return nil
}
