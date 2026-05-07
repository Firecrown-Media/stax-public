package files

import (
	"testing"

	"github.com/firecrown-media/stax/pkg/config"
)

func TestBuildSyncOptions_Defaults(t *testing.T) {
	cfg := config.Defaults()
	opts := BuildSyncOptions(cfg, SyncFlags{})

	if opts.DryRun {
		t.Error("expected DryRun false by default")
	}
	if opts.Delete {
		t.Error("expected Delete false by default")
	}
}

func TestBuildSyncOptions_BandwidthFromConfig(t *testing.T) {
	cfg := config.Defaults()
	cfg.Performance.RsyncBandwidthLimit = 500

	opts := BuildSyncOptions(cfg, SyncFlags{BandwidthLimit: 0})
	if opts.BandwidthLimit != 500 {
		t.Errorf("expected bandwidth 500 from config, got %d", opts.BandwidthLimit)
	}
}

func TestBuildSyncOptions_FlagOverridesConfig(t *testing.T) {
	cfg := config.Defaults()
	cfg.Performance.RsyncBandwidthLimit = 500

	opts := BuildSyncOptions(cfg, SyncFlags{BandwidthLimit: 1000})
	if opts.BandwidthLimit != 1000 {
		t.Errorf("expected flag value 1000 to override config 500, got %d", opts.BandwidthLimit)
	}
}
