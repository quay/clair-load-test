package manifests

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"

	"github.com/quay/zlog"
)

// execClairCtl executes the clairctl manifest command to fetch manifest.
// It returns the manifest bytes and errors if any during the execution.
func execClairCtl(ctx context.Context, container string) ([]byte, error) {
	cmd := exec.Command("clairctl", "manifest", container)
	zlog.Debug(ctx).Str("container", cmd.String()).Msg("getting manifest")
	return cmd.Output()
}

// batchProcess uses multiprocessing to get manifests and manifestHashes for a batch of containers.
// It returns a lists of manifests and manifestHashes.
func batchProcess(ctx context.Context, containers []string) ([][]byte, []string) {
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
				results <- result{container: cc, err: fmt.Errorf("could not generate manifest: %w", err)}
				return
			}
			err = json.Unmarshal(manifest, &blob)
			if err != nil {
				results <- result{container: cc, err: fmt.Errorf("could not extract hash from manifest: %w", err)}
				return
			}
			mu.Lock()
			defer mu.Unlock()
			listOfManifests[i] = manifest
			listOfManifestHashes[i] = blob.ManifestHash
			results <- result{}
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

// GetManifest uses multiprocessing to get manifests and manifestHashes for a list of containers.
// It returns a lists of manifests and manifestHashes.
func GetManifest(ctx context.Context, containers []string, concurrency int) ([][]byte, []string) {
	listOfManifests := make([][]byte, 0)
	updatedListOfManifests := make([][]byte, 0)
	listOfManifestHashes := make([]string, 0)
	updatedListOfManifestHashes := make([]string, 0)
	// Process containers in batches
	for i := 0; i < len(containers); i += concurrency {
		end := i + concurrency
		if end > len(containers) {
			end = len(containers)
		}
		batch := containers[i:end]
		manifestsBatch, manifestHashesBatch := batchProcess(ctx, batch)
		listOfManifests = append(listOfManifests, manifestsBatch...)
		listOfManifestHashes = append(listOfManifestHashes, manifestHashesBatch...)
	}
	for _, subList := range listOfManifests {
		if len(subList) > 0 {
			updatedListOfManifests = append(updatedListOfManifests, subList)
		}
	}
	for _, str := range listOfManifestHashes {
		if str != "" {
			updatedListOfManifestHashes = append(updatedListOfManifestHashes, str)
		}
	}
	return updatedListOfManifests, updatedListOfManifestHashes
}
