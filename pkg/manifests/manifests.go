package manifests

import (
	"sync"
	"context"
	"os/exec"
	"encoding/json"
	"github.com/quay/zlog"
)

func execClairCtl(ctx context.Context, container string) ([]byte, error) {
	cmd := exec.Command("clairctl", "manifest", container)
	zlog.Debug(ctx).Str("container", cmd.String()).Msg("getting manifest")
	return cmd.Output()
}

func getManifest(ctx context.Context, containers []string) []string {
	var blob ManifestHash
	var wg sync.WaitGroup
	var mu sync.Mutex
	listOfManifests := make([]string, len(containers))
	for i := 0; i < len(containers); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cc := containers[i]
			manifest, err := execClairCtl(ctx, cc)
			if err != nil {
				zlog.Debug(ctx).Str("container", cc).Msg(err.Error())
				return
			}
			err = json.Unmarshal(manifest, &blob)
			if err != nil {
				zlog.Error(ctx).Str("container", cc).Msg(err.Error())
				return
			}
			mu.Lock()
			defer mu.Unlock()
			listOfManifests[i] = blob.ManifestHash
		}(i)
	}
	wg.Wait()
	return listOfManifests
}

