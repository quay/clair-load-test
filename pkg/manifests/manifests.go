package manifests

import (
	"fmt"
	"sync"
	"context"
	"os/exec"
	"encoding/json"
	"github.com/quay/zlog"
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
	for i := 0; i < len(containers); i++ {
		wg.Add(1)
		go func(i int) error {
			defer wg.Done()
			cc := containers[i]
			manifest, err := execClairCtl(ctx, cc)
			if err != nil {
				return fmt.Errorf("could not generate manifest: %w", err)
			}
			err = json.Unmarshal(manifest, &blob)
			if err != nil {
				return fmt.Errorf("could not extract hash from manifest: %w", err)
			}
			mu.Lock()
			defer mu.Unlock()
			listOfManifests[i] = manifest
			listOfManifestHashes[i] = blob.ManifestHash
			return nil
		}(i)
	}
	wg.Wait()
	return listOfManifests, listOfManifestHashes
}

