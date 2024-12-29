package traefik_plugin_servefile

import (
	"context"
	"net/http"
	"os"
	"path"
	"strings"
)

type Config struct {
	RootDir      string `json:"rootDir,omitempty"`
	FallbackFile string `json:"fallbackFile,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		RootDir:      "",
		FallbackFile: "",
	}
}

type servefile struct {
	name         string
	next         http.Handler
	rootDir      string
	fallbackFile string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &servefile{
		name:         name,
		next:         next,
		rootDir:      config.RootDir,
		fallbackFile: config.FallbackFile,
	}, nil
}

func PathExists(path string) bool {
	_, err := os.Stat(path)

	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

func (b *servefile) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rootDir := b.rootDir
	reqPath := req.URL.Path
	if rootDir == "" || strings.Contains(reqPath, "..") {
		rw.WriteHeader(403)
		return
	}

	filePath := path.Join(rootDir, reqPath)
	if !PathExists(filePath) {
		fallbackFile := b.fallbackFile
		if fallbackFile != "" {
			filePath = path.Join(rootDir, fallbackFile)
		}
	}

	rw.Header().Del("Content-Type")
	http.ServeFile(rw, req, filePath)
}
