/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package data

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/svc-media/biz"
)

type localStorage struct {
	baseDir string
	log     *log.Helper
}

// NewLocalStorage .
func NewLocalStorage(baseDir string, logger log.Logger) biz.Storage {
	return &localStorage{
		baseDir: baseDir,
		log:     log.NewHelper(logger),
	}
}

func (s *localStorage) StorePart(ctx context.Context, uploadID string, partNumber int, data []byte) (string, error) {
	tempPath := filepath.Join(s.baseDir, "temp", uploadID)
	if err := os.MkdirAll(tempPath, 0755); err != nil {
		return "", err
	}

	partFilename := fmt.Sprintf("part_%d", partNumber)
	partFilePath := filepath.Join(tempPath, partFilename)

	if err := os.WriteFile(partFilePath, data, 0644); err != nil {
		return "", err
	}

	// For local storage, we can use the part number as an etag for simplicity.
	return fmt.Sprintf("part_%d_etag", partNumber), nil
}

func (s *localStorage) MergeParts(ctx context.Context, uploadID string, totalParts int, finalPath string) error {
	tempPath := filepath.Join(s.baseDir, "temp", uploadID)
	finalFilePath := filepath.Join(s.baseDir, finalPath)

	if err := os.MkdirAll(filepath.Dir(finalFilePath), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(finalFilePath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	for i := 1; i <= totalParts; i++ {
		partFilename := fmt.Sprintf("part_%d", i)
		partFilePath := filepath.Join(tempPath, partFilename)

		partFile, err := os.Open(partFilePath)
		if err != nil {
			return fmt.Errorf("failed to open part %d: %w", i, err)
		}

		if _, err := io.Copy(destFile, partFile); err != nil {
			partFile.Close()
			return fmt.Errorf("failed to copy part %d: %w", i, err)
		}
		partFile.Close()
	}

	return nil
}

func (s *localStorage) DeleteParts(ctx context.Context, uploadID string) error {
	tempPath := filepath.Join(s.baseDir, "temp", uploadID)
	return os.RemoveAll(tempPath)
}
