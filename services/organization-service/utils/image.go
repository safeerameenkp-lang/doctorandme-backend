package utils

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxFileSize    = 5 * 1024 * 1024 // 5MB
	AllowedFormats = ".jpg,.jpeg,.png"
)

// ValidateImage checks file size and format
func ValidateImage(fileHeader *multipart.FileHeader) error {
	if fileHeader.Size > MaxFileSize {
		return errors.New("file size exceeds 5MB limit")
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return nil
	default:
		return errors.New("unsupported file format. Allowed formats: JPG, JPEG, PNG")
	}
}

// SaveOptimizedImage saves the uploaded image after optimization (compression)
// Returns the relative path to the saved image
func SaveOptimizedImage(file multipart.File, originalFilename string, subDir string) (string, error) {
	// Create upload directory if not exists
	uploadPath := filepath.Join("uploads", subDir)
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Decode image to check validity and prepare for re-encoding
	// We need to decode first to determine the format and content
	// Since file is multipart.File, it should be a Seeker
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to seek file: %w", err)
	}

	img, format, err := image.Decode(file)
	if err != nil {
		return "", errors.New("invalid image content")
	}

	// Determine extension based on actual format
	ext := "jpg"
	if format == "png" {
		ext = "png"
	}
	// default to jpg for jpeg

	// Generate unique filename
	newFilename := fmt.Sprintf("%s_%d.%s", uuid.New().String(), time.Now().Unix(), ext)

	// Create destination file
	dstPath := filepath.Join(uploadPath, newFilename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Re-encode with optimization
	switch format {
	case "jpeg", "jpg":
		// Encode as JPEG with quality 80
		err = jpeg.Encode(dst, img, &jpeg.Options{Quality: 80})
	case "png":
		// Encode as PNG
		err = png.Encode(dst, img)
	default:
		// Should not happen if Decode succeeded and we only support common formats,
		// unless image library supports others not in our allowed list.
		// For safety, force encode as JPEG if unknown but decodeable?
		// Or return error.
		// ValidateImage already checked extension, so it's likely one of them.
		// If it's gif (standard lib supports gif decoding), we might end up here.
		// We can convert gif to jpeg.
		err = jpeg.Encode(dst, img, &jpeg.Options{Quality: 80})
	}

	if err != nil {
		return "", fmt.Errorf("failed to save optimized image: %w", err)
	}

	return dstPath, nil
}
