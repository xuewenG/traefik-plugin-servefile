package traefik_plugin_servefile

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path"
	"strings"
)

type Config struct {
	RootDir      string `json:"rootDir,omitempty"`
	DefaultFile string `json:"defaultFile,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		RootDir:      "",
		DefaultFile: "",
	}
}

type servefile struct {
	name         string
	next         http.Handler
	rootDir      string
	defaultFile string
}

func exists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.IsDir()
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	rootDir := config.RootDir
	if rootDir == "" {
		return nil, errors.New("rootDir is required")
	}

	if !exists(rootDir) {
		return nil, errors.New("rootDir does not exist")
	}

	if !isDir(rootDir) {
		return nil, errors.New("rootDir is not a directory")
	}

	defaultFile := config.DefaultFile
	if defaultFile != "" {
		defaultFile = path.Join(config.RootDir, defaultFile)
		if !exists(defaultFile) {
			return nil, errors.New("defaultFile does not exist")
		}
	}

	return &servefile{
		name:         name,
		next:         next,
		rootDir:      rootDir,
		defaultFile: defaultFile,
	}, nil
}

func (b *servefile) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Del("Content-Type")

	reqPath := req.URL.Path
	if strings.Contains(reqPath, "..") {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	filePath := path.Join(b.rootDir, reqPath)
	if !exists(filePath) && b.defaultFile != "" {
		filePath = path.Join(b.rootDir, b.defaultFile)
	}

	http.ServeFile(rw, req, filePath)
}
