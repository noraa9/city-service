package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage saves files on the server filesystem and returns a public URL.
//
// This is very convenient for platforms like Railway where attaching a separate MinIO
// service might be overkill for early testing.
//
// IMPORTANT: Local disk on many PaaS providers is ephemeral (can be wiped on redeploy).
// Use MinIO/S3 for production durability.
type LocalStorage struct {
	dir          string // e.g. "./uploads"
	publicBase   string // e.g. "https://myapp.up.railway.app"
	publicPrefix string // e.g. "/uploads"
}

func NewLocalStorage(dir, publicBaseURL, publicPrefix string) (*LocalStorage, error) {
	if dir == "" {
		return nil, fmt.Errorf("local storage dir is required")
	}
	if publicPrefix == "" {
		publicPrefix = "/uploads"
	}
	publicPrefix = "/" + strings.Trim(publicPrefix, "/")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create uploads dir: %w", err)
	}

	return &LocalStorage{
		dir:          dir,
		publicBase:   strings.TrimRight(publicBaseURL, "/"),
		publicPrefix: publicPrefix,
	}, nil
}

func (s *LocalStorage) UploadFile(ctx context.Context, bucketName, fileName string, file io.Reader, size int64, contentType string) (string, error) {
	_ = ctx
	_ = bucketName
	_ = size
	_ = contentType

	// fileName should already be unique (uuid + ext). We still clean path separators.
	fileName = filepath.Base(fileName)
	dstPath := filepath.Join(s.dir, fileName)

	f, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, file); err != nil {
		return "", fmt.Errorf("save file: %w", err)
	}

	escaped := url.PathEscape(fileName)
	if s.publicBase == "" {
		// If PUBLIC_BASE_URL not set, return a path. Handlers/UI can still display it relative to host.
		return fmt.Sprintf("%s/%s", s.publicPrefix, escaped), nil
	}
	return fmt.Sprintf("%s%s/%s", s.publicBase, s.publicPrefix, escaped), nil
}
