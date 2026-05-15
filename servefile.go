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
	RootDir     string `json:"rootDir,omitempty"`
	DefaultFile string `json:"defaultFile,omitempty"`
	ListDir     bool   `json:"listDir,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		RootDir:     "",
		DefaultFile: "",
		ListDir:     false,
	}
}

type servefile struct {
	name        string
	next        http.Handler
	rootDir     string
	defaultFile string
	listDir     bool
}

func exists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.IsDir()
}

func ensureEndsWithSlash(path string) string {
	if strings.HasSuffix(path, "/") {
		return path
	}

	return path + "/"
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	rootDir := config.RootDir
	if rootDir == "" {
		return nil, errors.New("rootDir is required")
	}

	rootDir = path.Clean(rootDir)
	if !exists(rootDir) {
		return nil, errors.New("rootDir does not exist")
	}

	if !isDir(rootDir) {
		return nil, errors.New("rootDir is not a directory")
	}

	defaultFile := config.DefaultFile
	if defaultFile != "" {
		defaultFile = path.Clean(path.Join(rootDir, config.DefaultFile))
		if !exists(defaultFile) {
			return nil, errors.New("defaultFile does not exist")
		}
	}

	return &servefile{
		name:        name,
		next:        next,
		rootDir:     rootDir,
		defaultFile: defaultFile,
		listDir:     config.ListDir,
	}, nil
}

func (b *servefile) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Del("Content-Type")

	reqPath := req.URL.Path
	if strings.Contains(reqPath, "..") {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	filePath := path.Clean(path.Join(b.rootDir, reqPath))

	if !exists(filePath) {
		if b.defaultFile == "" {
			rw.WriteHeader(http.StatusForbidden)
			return
		}

		filePath = b.defaultFile
	}

	if isDir(filePath) && !b.listDir {
		if b.defaultFile == "" {
			rw.WriteHeader(http.StatusForbidden)
			return
		}

		filePath = b.defaultFile
	}

	if !strings.HasPrefix(filePath, ensureEndsWithSlash(b.rootDir)) {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	http.ServeFile(rw, req, filePath)
}
