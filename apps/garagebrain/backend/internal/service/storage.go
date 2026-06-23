package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Storage сохраняет файлы (фото чеков) на локальный диск VPS.
// Для MVP не используется S3/R2: файлы лежат под UPLOAD_DIR и раздаются
// nginx'ом (location /uploads) либо самим Go-сервером в dev-режиме.
type Storage struct {
	baseDir   string // абсолютный/относительный путь на диске, напр. /var/www/garagebrain/uploads
	publicURL string // префикс URL, под которым файлы доступны, напр. /uploads
}

func NewStorage() *Storage {
	baseDir := os.Getenv("UPLOAD_DIR")
	if baseDir == "" {
		baseDir = "./uploads"
	}
	publicURL := os.Getenv("UPLOAD_PUBLIC_URL")
	if publicURL == "" {
		publicURL = "/uploads"
	}
	return &Storage{baseDir: baseDir, publicURL: strings.TrimRight(publicURL, "/")}
}

// BaseDir — корневая директория хранилища (для статической раздачи).
func (s *Storage) BaseDir() string { return s.baseDir }

// PublicPrefix — URL-префикс хранилища.
func (s *Storage) PublicPrefix() string { return s.publicURL }

var allowedImageExt = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".heic": "image/heic",
}

// Save записывает содержимое файла в UPLOAD_DIR/<carID>/<uuid><ext> и
// возвращает публичный URL. ext берётся из исходного имени (по умолчанию .jpg).
func (s *Storage) Save(carID uuid.UUID, filename string, src io.Reader) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if _, ok := allowedImageExt[ext]; !ok {
		ext = ".jpg"
	}

	dir := filepath.Join(s.baseDir, carID.String())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	name := uuid.NewString() + ext
	dst := filepath.Join(dir, name)

	f, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, src); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return fmt.Sprintf("%s/%s/%s", s.publicURL, carID.String(), name), nil
}

// IsAllowedImage сообщает, поддерживается ли расширение файла.
func IsAllowedImage(filename string) bool {
	_, ok := allowedImageExt[strings.ToLower(filepath.Ext(filename))]
	return ok
}
