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

// ReadFile reads a CBZ file and returns its contents
func ReadFile(filename string) (*File, error) {
	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CBZ file: %w", err)
	}
	defer zipReader.Close()

	cbzFile := &File{
		Name:   filename,
		Images: []Image{},
	}

	// Read all image files from the zip
	for _, file := range zipReader.File {
		// Skip directories and non-image files
		if file.FileInfo().IsDir() || !isImageFile(file.Name) {
			continue
		}

		// Open the file inside the zip
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file in CBZ: %w", err)
		}

		// Read the file data
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read file data: %w", err)
		}

		// Add the image to the CBZ file
		cbzFile.Images = append(cbzFile.Images, Image{
			Name:     filepath.Base(file.Name),
			Data:     data,
			MimeType: getMimeType(file.Name),
		})
	}

	// Sort images by name
	sort.Slice(cbzFile.Images, func(i, j int) bool {
		return cbzFile.Images[i].Name < cbzFile.Images[j].Name
	})

	return cbzFile, nil
}

// MergeFiles merges multiple CBZ files into one
func MergeFiles(inputFiles []string, outputFile string) error {
	// Create a new zip file
	zipFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Process each input file
	imageCounter := 1
	for chapterIndex, inputFile := range inputFiles {
		cbzFile, err := ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file %s: %w", inputFile, err)
		}

		// Add each image to the output zip with a new name to avoid conflicts
		for _, image := range cbzFile.Images {
			// Create a new name for the image: chapterXXX_imageYYY.ext
			ext := filepath.Ext(image.Name)
			newName := fmt.Sprintf("chapter%03d_%03d%s", chapterIndex+1, imageCounter, ext)
			imageCounter++

			// Create a new file in the zip
			writer, err := zipWriter.Create(newName)
			if err != nil {
				return fmt.Errorf("failed to create file in output zip: %w", err)
			}

			// Write the image data
			_, err = writer.Write(image.Data)
			if err != nil {
				return fmt.Errorf("failed to write image data: %w", err)
			}
		}
	}

	return nil
}

// isImageFile checks if a file is an image based on its extension
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"
}

// getMimeType returns the MIME type for a file based on its extension
func getMimeType(filename string) string {
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
