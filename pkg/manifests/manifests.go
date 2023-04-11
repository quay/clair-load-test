package manifests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/quay/zlog"
	"os/exec"
	"sync"
)

// Method to execute clairctl manifest command.
func execClairCtl(ctx context.Context, container string) ([]byte, error) {
	cmd := exec.Command("clairctl", "manifest", container)
	zlog.Debug(ctx).Str("container", cmd.String()).Msg("getting manifest")
	return cmd.Output()
}

// Method to process containers and extract their manifests.
func GetManifest(ctx context.Context, containers []string) ([][]byte, []string) {
	var blob ManifestHash
	var wg sync.WaitGroup
	var mu sync.Mutex
	listOfManifests := make([][]byte, len(containers))
	listOfManifestHashes := make([]string, len(containers))
	results := make(chan result)
	for i := 0; i < len(containers); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cc := containers[i]
			manifest, err := execClairCtl(ctx, cc)
			if err != nil {
				results <- result{index: i, container: cc, err: fmt.Errorf("could not generate manifest: %w", err)}
				return
			}
			err = json.Unmarshal(manifest, &blob)
			if err != nil {
				results <- result{index: i, container: cc, err: fmt.Errorf("could not extract hash from manifest: %w", err)}
				return
			}
			mu.Lock()
			defer mu.Unlock()
			listOfManifests[i] = manifest
			listOfManifestHashes[i] = blob.ManifestHash
			results <- result{index: i}
		}(i)
	}
	go func() {
		for res := range results {
			if res.err != nil {
				zlog.Debug(ctx).Str("container", res.container).Msg(fmt.Sprintf("Error generating manifest for container. Message: %v", res.err))
			}
		}
	}()
	wg.Wait()
	close(results)
	return listOfManifests, listOfManifestHashes
}
