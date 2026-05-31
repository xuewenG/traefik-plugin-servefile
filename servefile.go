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
	IndexFile    string `json:"indexFile,omitempty"`
	FallbackFile string `json:"fallbackFile,omitempty"`
	ListDir      bool   `json:"listDir,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

type servefile struct {
	name string
	next http.Handler

	rootDir      string
	indexFile    string
	fallbackFile string
	listDir      bool
}

func exists(filePath string) bool {
	_, err := os.Stat(filePath)

	return err == nil
}

func isDir(filePath string) bool {
	info, err := os.Stat(filePath)

	return err == nil && info.IsDir()
}

func isFile(filePath string) bool {
	info, err := os.Stat(filePath)

	return err == nil && !info.IsDir()
}

func ensureEndsWithSlash(filePath string) string {
	if strings.HasSuffix(filePath, "/") {
		return filePath
	}

	return filePath + "/"
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	rootDir := config.RootDir
	if rootDir == "" {
		return nil, errors.New("rootDir is required")
	}

	rootDir = path.Clean(rootDir)
	if !isDir(rootDir) {
		return nil, errors.New("rootDir is not a directory")
	}

	indexFile := config.IndexFile
	if indexFile == "" {
		indexFile = "index.html"
	}

	return &servefile{
		name: name,
		next: next,

		rootDir:      rootDir,
		indexFile:    indexFile,
		fallbackFile: config.FallbackFile,
		listDir:      config.ListDir,
	}, nil
}

func (b *servefile) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Del("Content-Type")

	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqPath := req.URL.Path
	if strings.Contains(reqPath, "..") {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	filePath := path.Join(b.rootDir, reqPath)

	if isDir(filePath) {
		indexPath := path.Join(filePath, b.indexFile)
		if isFile(indexPath) {
			filePath = indexPath
		}
	}

	if !exists(filePath) || (!b.listDir && isDir(filePath)) {
		if b.fallbackFile == "" {
			rw.WriteHeader(http.StatusForbidden)
			return
		}

		fallbackFilePath := path.Join(b.rootDir, b.fallbackFile)
		if !isFile(fallbackFilePath) {
			rw.WriteHeader(http.StatusForbidden)
			return
		}

		filePath = fallbackFilePath
	}

	if filePath != b.rootDir && !strings.HasPrefix(filePath, ensureEndsWithSlash(b.rootDir)) {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	http.ServeFile(rw, req, filePath)
}
