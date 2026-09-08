package cbz

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// File represents a CBZ file with its contents
type File struct {
	Name   string
	Images []Image
}

// Image represents an image inside a CBZ file
type Image struct {
	Name     string
	Data     []byte
	MimeType string
}

// ImageHandler is a callback that processes one image streamed from a CBZ
// archive. It receives the base name, an io.Reader positioned at the image
// bytes, and the MIME type. The handler must consume the reader before it
// returns; returning an error stops iteration and propagates the error.
type ImageHandler func(name string, data io.Reader, mimeType string) error

// IterateImages opens a CBZ file and invokes handler for each image in sorted
// order without buffering the whole archive in memory. Prefer this over
// ReadFile when the caller only needs to stream images through to an output.
func IterateImages(filename string, handler ImageHandler) error {
	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		return fmt.Errorf("failed to open CBZ file: %w", err)
	}
	defer zipReader.Close()

	// Collect image entries, skipping directories and non-images.
	var imageFiles []*zip.File
	for _, file := range zipReader.File {
		if !file.FileInfo().IsDir() && IsImageFile(file.Name) {
			imageFiles = append(imageFiles, file)
		}
	}

	// Process images in sorted order by base name.
	sort.Slice(imageFiles, func(i, j int) bool {
		return filepath.Base(imageFiles[i].Name) < filepath.Base(imageFiles[j].Name)
	})

	for _, file := range imageFiles {
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in CBZ: %w", err)
		}

		err = handler(filepath.Base(file.Name), rc, MimeType(file.Name))
		rc.Close()
		if err != nil {
			return fmt.Errorf("error processing image %s: %w", file.Name, err)
		}
	}

	return nil
}

// ReadFile reads a CBZ file and returns its contents. It loads the entire
// archive into memory; prefer IterateImages for streaming large files.
func ReadFile(filename string) (*File, error) {
	cbzFile := &File{
		Name:   filename,
		Images: []Image{},
	}

	err := IterateImages(filename, func(name string, data io.Reader, mimeType string) error {
		content, err := io.ReadAll(data)
		if err != nil {
			return fmt.Errorf("failed to read file data: %w", err)
		}
		cbzFile.Images = append(cbzFile.Images, Image{
			Name:     name,
			Data:     content,
			MimeType: mimeType,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return cbzFile, nil
}

// MergeFiles merges multiple CBZ files into one, streaming images to avoid
// loading all into memory at once.
func MergeFiles(inputFiles []string, outputFile string) error {
	// Create a new zip file
	zipFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Stream each input file's images straight into the output archive so we
	// never hold more than one image in memory at a time.
	imageCounter := 1
	for chapterIndex, inputFile := range inputFiles {
		err := IterateImages(inputFile, func(name string, data io.Reader, mimeType string) error {
			// Create a new name for the image: chapterXXX_imageYYY.ext
			ext := filepath.Ext(name)
			newName := fmt.Sprintf("chapter%03d_%03d%s", chapterIndex+1, imageCounter, ext)
			imageCounter++

			writer, err := zipWriter.Create(newName)
			if err != nil {
				return fmt.Errorf("failed to create file in output zip: %w", err)
			}

			if _, err := io.Copy(writer, data); err != nil {
				return fmt.Errorf("failed to write image data: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to read input file %s: %w", inputFile, err)
		}
	}

	return nil
}

// IsImageFile reports whether filename has a supported image extension.
func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"
}

// MimeType returns the MIME type for filename based on its extension.
func MimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
