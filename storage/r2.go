package storage

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Storage struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewR2Storage() (*R2Storage, error) {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucket := os.Getenv("R2_BUCKET_NAME")
	publicURL := os.Getenv("R2_PUBLIC_URL")

	if accountID == "" || accessKey == "" || secretKey == "" || bucket == "" || publicURL == "" {
		return nil, fmt.Errorf("missing R2 environment variables")
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	client := s3.New(s3.Options{
		EndpointResolver: s3.EndpointResolverFromURL(endpoint),
		Region:           "auto",
		Credentials:      credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
	})

	return &R2Storage{
		client:    client,
		bucket:    bucket,
		publicURL: publicURL,
	}, nil
}

func (r *R2Storage) UploadImage(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Valida tipo de arquivo
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
	}

	contentType, ok := allowed[ext]
	if !ok {
		return "", fmt.Errorf("file type not allowed, use jpg, png or webp")
	}

	// Valida tamanho (max 5MB)
	if header.Size > 5*1024*1024 {
		return "", fmt.Errorf("file too large, max 5MB")
	}

	// Gera nome único
	key := fmt.Sprintf("products/%d%s", time.Now().UnixNano(), ext)

	_, err := r.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	// Retorna URL pública
	url := fmt.Sprintf("%s/%s", strings.TrimRight(r.publicURL, "/"), key)
	return url, nil
}

// UploadImageBytes faz upload de bytes já processados
func (r *R2Storage) UploadImageBytes(key string, data []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucket), // MUDE para r.bucket
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

	if err != nil {
		return "", fmt.Errorf("erro ao fazer upload para R2: %w", err)
	}

	publicURL := fmt.Sprintf("%s/%s", r.publicURL, key)
	return publicURL, nil
}
