package traefik_plugin_servefile_test

import (
	"testing"

	traefik_plugin_servefile "github.com/xuewenG/traefik-plugin-servefile"
)

type HeaderConfig struct {
	Name  string
	Value string
}

func TestCreateConfig(t *testing.T) {
	t.Run("create config", func(t *testing.T) {
		config := traefik_plugin_servefile.CreateConfig()
		if config == nil {
			t.Fatal("create config failed")
		}
	})
}

func TestNew(t *testing.T) {
}

func TestServeHTTP(t *testing.T) {
}
