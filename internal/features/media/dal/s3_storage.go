/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/conf"
	"origadmin/application/origcms/internal/data/enums"
)

// S3Storage implements the Storage interface using S3-compatible object storage.
// It supports AWS S3, MinIO, and any S3-compatible service.
type S3Storage struct {
	client    *s3.Client
	presigner *s3.PresignClient
	uploader  *manager.Uploader
	bucket    string
	presign   time.Duration
	logger    *log.Helper
}

// NewS3Storage creates a new S3Storage instance from the given S3Config.
func NewS3Storage(cfg *conf.S3Config, logger log.Logger) (*S3Storage, error) {
	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(creds),
	}

	if cfg.Endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: cfg.Endpoint, SigningRegion: region}, nil
		})
		opts = append(opts, config.WithEndpointResolverWithOptions(customResolver))
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})

	presignExpiry := cfg.PresignExpiry
	if presignExpiry <= 0 {
		presignExpiry = 15 * time.Minute
	}

	return &S3Storage{
		client:    client,
		presigner: s3.NewPresignClient(client),
		uploader:  manager.NewUploader(client),
		bucket:    cfg.Bucket,
		presign:   presignExpiry,
		logger:    log.NewHelper(log.With(logger, "module", "storage.s3")),
	}, nil
}

// Upload uploads a file to S3 and returns the object key.
func (s *S3Storage) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   r,
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	_, err := s.uploader.Upload(ctx, input)
	if err != nil {
		return "", fmt.Errorf("S3 upload key=%s: %w", key, err)
	}

	return key, nil
}

// Download downloads a file from S3 by key.
func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("S3 download key=%s: %w", key, err)
	}
	return output.Body, nil
}

// Delete removes an object from S3 by key.
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("S3 delete key=%s: %w", key, err)
	}
	return nil
}

// GetURL returns a presigned URL for the given key.
func (s *S3Storage) GetURL(ctx context.Context, key string) (string, error) {
	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = s.presign
	})
	if err != nil {
		return "", fmt.Errorf("S3 presign key=%s: %w", key, err)
	}
	return req.URL, nil
}

// StorePart stores a single upload part in S3.
// For S3, we store parts as individual objects under a parts prefix.
func (s *S3Storage) StorePart(ctx context.Context, uploadID string, partNumber int, data []byte) (string, error) {
	key := fmt.Sprintf("temp/parts/%s/part_%05d", uploadID, partNumber)
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytesReader(data),
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("S3 store part uploadID=%s part=%d: %w", uploadID, partNumber, err)
	}

	// Return ETag-like identifier
	return fmt.Sprintf("%s:%d", key, partNumber), nil
}

// MergeParts merges all parts into a single file in S3.
// For S3, we download all parts and upload the concatenated result.
func (s *S3Storage) MergeParts(ctx context.Context, uploadID string, totalParts int, finalPath string) error {
	// Use multipart copy to assemble parts into the final object
	// For simplicity, we use a single Upload with concatenated data
	// A production implementation would use S3 multipart upload with CopyPart

	// Create a pipe to stream merged data
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		for i := 1; i <= totalParts; i++ {
			partKey := fmt.Sprintf("temp/parts/%s/part_%05d", uploadID, i)
			output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(s.bucket),
				Key:    aws.String(partKey),
			})
			if err != nil {
				pw.CloseWithError(fmt.Errorf("read part %d: %w", i, err))
				return
			}
			if _, err := io.Copy(pw, output.Body); err != nil {
				output.Body.Close()
				pw.CloseWithError(fmt.Errorf("copy part %d: %w", i, err))
				return
			}
			output.Body.Close()
		}
	}()

	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(finalPath),
		Body:   pr,
	})
	if err != nil {
		return fmt.Errorf("S3 merge parts uploadID=%s: %w", uploadID, err)
	}

	// Clean up part objects
	for i := 1; i <= totalParts; i++ {
		partKey := fmt.Sprintf("temp/parts/%s/part_%05d", uploadID, i)
		_, _ = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(partKey),
		})
	}

	return nil
}

// DeleteParts removes all part objects for an upload session from S3.
func (s *S3Storage) DeleteParts(ctx context.Context, uploadID string) error {
	prefix := fmt.Sprintf("temp/parts/%s/", uploadID)

	listOutput, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return fmt.Errorf("S3 list parts uploadID=%s: %w", uploadID, err)
	}

	for _, obj := range listOutput.Contents {
		_, _ = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    obj.Key,
		})
	}

	return nil
}

// PromoteToOriginal copies a file from temp/ to originals/ in S3.
func (s *S3Storage) PromoteToOriginal(ctx context.Context, tempPath string) (string, error) {
	// Parse temp path to construct originals path
	// Expected format: temp/{userID}/{yyyy}/{MM}/{filename}
	// Target format:   originals/{userID}/{yyyy}/{MM}/{filename}
	if len(tempPath) < 6 || tempPath[:5] != "temp/" {
		return "", fmt.Errorf("invalid temp path format: %s", tempPath)
	}
	originalPath := "originals/" + tempPath[5:]

	// Copy object from temp to originals
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(originalPath),
		CopySource: aws.String(s.bucket + "/" + tempPath),
	})
	if err != nil {
		return "", fmt.Errorf("S3 copy temp→originals: %w", err)
	}

	// Delete the temp object after successful copy
	_, _ = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(tempPath),
	})

	return originalPath, nil
}

// CleanupTempParts removes the parts directory for an upload session from S3.
func (s *S3Storage) CleanupTempParts(ctx context.Context, userID, uploadID string) error {
	// S3 does not have directories; parts are already cleaned in DeleteParts/MergeParts.
	// This is a no-op for S3 storage.
	return nil
}

// SyncStatus returns the sync status for a key in S3.
// For S3-only storage, files are always synced (they are in S3).
func (s *S3Storage) SyncStatus(ctx context.Context, key string) (enums.SyncStatus, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return enums.SyncStatusFailed, nil
	}
	return enums.SyncStatusSynced, nil
}

// bytesReader creates an io.Reader from a byte slice.
type bytesReaderWrapper struct {
	data []byte
	pos  int
}

func bytesReader(data []byte) *bytesReaderWrapper {
	return &bytesReaderWrapper{data: data}
}

func (r *bytesReaderWrapper) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
