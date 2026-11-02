package config

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"testing"
)

func TestDeviceAddressBooksDeploymentConfig(t *testing.T) {
	for _, tc := range []struct {
		name, yaml, env string
		want            bool
	}{
		{"missing", "app: {}\n", "", false},
		{"disabled", "app:\n  device-address-books: false\n", "", false},
		{"enabled", "app:\n  device-address-books: true\n", "", true},
		{"environment only", "app: {}\n", "true", true},
		{"environment disables yaml", "app:\n  device-address-books: true\n", "false", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			t.Setenv("RUSTDESK_API_APP_DEVICE_ADDRESS_BOOKS", tc.env)
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(tc.yaml), 0600); err != nil {
				t.Fatal(err)
			}
			var c Config
			Init(&c, path)
			if c.App.DeviceAddressBooks != tc.want {
				t.Fatalf("got %v, want %v", c.App.DeviceAddressBooks, tc.want)
			}
		})
	}
}
