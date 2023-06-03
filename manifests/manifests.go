package manifests

import (
	"context"
	"encoding/json"
	"os/exec"
	"sync"

	"github.com/quay/zlog"
	"golang.org/x/sync/errgroup"
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
	var mu sync.Mutex
	listOfManifests := make([][]byte, len(containers))
	listOfManifestHashes := make([]string, len(containers))
	errGroup, ctx := errgroup.WithContext(ctx)

	for idx := 0; idx < len(containers); idx++ {
		idx := idx // Capture loop variable
		errGroup.Go(func() error {
			cc := containers[idx]
			manifest, err := execClairCtl(ctx, cc)
			if err != nil {
				zlog.Debug(ctx).Str("container", cc).Msg("Could not generate manifest")
			}
			err = json.Unmarshal(manifest, &blob)
			if err != nil {
				zlog.Debug(ctx).Str("container", cc).Msg("Could not extract hash from manifest JSON")
			}
			mu.Lock()
			defer mu.Unlock()
			listOfManifests[idx] = manifest
			listOfManifestHashes[idx] = blob.ManifestHash
			return nil
		})
	}
	_ = errGroup.Wait()
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
