package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/imovxen/backend/internal/config"
)

type StorageService struct {
	cfg *config.Config
}

type PresignResponse struct {
	UploadURL string `json:"upload_url"`
	FileURL   string `json:"file_url"`
	Key       string `json:"key"`
}

func NewStorageService(cfg *config.Config) *StorageService {
	return &StorageService{cfg: cfg}
}

// GeneratePresignedUploadURL cria uma URL de upload PUT simulada/compatível para S3/MinIO
func (s *StorageService) GeneratePresignedUploadURL(ctx context.Context, propertyID uuid.UUID, filename string, mimeType string) (*PresignResponse, error) {
	allowedMimes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	if !allowedMimes[strings.ToLower(mimeType)] {
		return nil, fmt.Errorf("formato de imagem inválido: %s. Permitidos: jpg, png, webp", mimeType)
	}

	ext := ".jpg"
	if strings.Contains(mimeType, "png") {
		ext = ".png"
	} else if strings.Contains(mimeType, "webp") {
		ext = ".webp"
	}

	fileKey := fmt.Sprintf("properties/%s/%s%s", propertyID.String(), uuid.New().String(), ext)
	fileURL := fmt.Sprintf("%s/%s", strings.TrimRight(s.cfg.S3PublicBaseURL, "/"), fileKey)
	uploadURL := fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.cfg.S3Endpoint, "/"), s.cfg.S3Bucket, fileKey)

	return &PresignResponse{
		UploadURL: uploadURL,
		FileURL:   fileURL,
		Key:       fileKey,
	}, nil
}
