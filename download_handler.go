package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	maxDownloadBytes = 50 << 20 // 50 MB
	downloadTimeout  = 30 * time.Second
)

type downloadRequest struct {
	URL string `json:"url"`
}

type downloadResponse struct {
	FileName string `json:"fileName"`
	Path     string `json:"path"`
	Bytes    int64  `json:"bytes"`
}

func NewDownloadHandler(downloadDir string) http.HandlerFunc {
	client := &http.Client{Timeout: downloadTimeout}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req downloadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		rawURL := strings.TrimSpace(req.URL)
		if rawURL == "" {
			http.Error(w, "url is required", http.StatusBadRequest)
			return
		}

		parsedURL, err := url.Parse(rawURL)
		if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			http.Error(w, "url must be a valid http or https URL", http.StatusBadRequest)
			return
		}

		if err := os.MkdirAll(downloadDir, 0o755); err != nil {
			http.Error(w, "failed to prepare download directory", http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), downloadTimeout)
		defer cancel()

		remoteReq, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
		if err != nil {
			http.Error(w, "failed to build download request", http.StatusInternalServerError)
			return
		}

		resp, err := client.Do(remoteReq)
		if err != nil {
			http.Error(w, "failed to download file", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			http.Error(w, fmt.Sprintf("download failed with status %d", resp.StatusCode), http.StatusBadGateway)
			return
		}

		pattern := tempFilePatternFromURL(parsedURL)
		file, err := os.CreateTemp(downloadDir, pattern)
		if err != nil {
			http.Error(w, "failed to create destination file", http.StatusInternalServerError)
			return
		}

		targetPath := file.Name()
		bytesWritten, copyErr := copyWithLimit(file, resp.Body, maxDownloadBytes)
		closeErr := file.Close()
		if copyErr != nil || closeErr != nil {
			os.Remove(targetPath)
			if errors.Is(copyErr, errDownloadTooLarge) {
				http.Error(w, "download exceeds max allowed size (50 MB)", http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, "failed to save downloaded file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(downloadResponse{
			FileName: filepath.Base(targetPath),
			Path:     targetPath,
			Bytes:    bytesWritten,
		})
	}
}

var errDownloadTooLarge = errors.New("download too large")

func copyWithLimit(dst io.Writer, src io.Reader, maxBytes int64) (int64, error) {
	written, err := io.Copy(dst, io.LimitReader(src, maxBytes+1))
	if err != nil {
		return written, err
	}
	if written > maxBytes {
		return written, errDownloadTooLarge
	}
	return written, nil
}

var fileNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func tempFilePatternFromURL(parsedURL *url.URL) string {
	base := path.Base(parsedURL.Path)
	if base == "." || base == "/" || base == "" {
		base = "downloaded-file.bin"
	}

	ext := path.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	stem = fileNameSanitizer.ReplaceAllString(stem, "_")
	if stem == "" {
		stem = "downloaded-file"
	}
	ext = fileNameSanitizer.ReplaceAllString(ext, "")
	if ext == "" {
		ext = ".bin"
	}

	return stem + "-*" + ext
}
