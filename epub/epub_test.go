package epub

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cbz2epub/cbz"
)

// createTestCBZ creates a test CBZ file with the given images
func createTestCBZ(t *testing.T, filename string, images []struct{ name, content string }) *cbz.File {
	// Create a new zip file
	zipFile, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create test CBZ file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add each image to the zip
	for _, image := range images {
		writer, err := zipWriter.Create(image.name)
		if err != nil {
			t.Fatalf("Failed to create file in test CBZ: %v", err)
		}

		_, err = writer.Write([]byte(image.content))
		if err != nil {
			t.Fatalf("Failed to write image data in test CBZ: %v", err)
		}
	}

	// Create a CBZ File object
	cbzFile := &cbz.File{
		Name:   filename,
		Images: []cbz.Image{},
	}

	// Add images to the CBZ File object
	for _, image := range images {
		// Skip non-image files
		if !isImageFile(image.name) {
			continue
		}

		cbzFile.Images = append(cbzFile.Images, cbz.Image{
			Name:     filepath.Base(image.name),
			Data:     []byte(image.content),
			MimeType: getMimeType(image.name),
		})
	}

	return cbzFile
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

// TestConvertFromCBZ tests the ConvertFromCBZ function
func TestConvertFromCBZ(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "epub_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test CBZ file object
	testImages := []struct{ name, content string }{
		{"image1.jpg", "test image 1 content"},
		{"image2.png", "test image 2 content"},
		{"subfolder/image3.gif", "test image 3 content"},
		{"not_an_image.txt", "this is not an image"},
	}

	testCBZPath := filepath.Join(tempDir, "test.cbz")
	cbzFile := createTestCBZ(t, testCBZPath, testImages)

	// Convert the CBZ to EPUB
	epubPath := filepath.Join(tempDir, "test.epub")
	err = ConvertFromCBZ(cbzFile, epubPath)
	if err != nil {
		t.Fatalf("ConvertFromCBZ failed: %v", err)
	}

	// Check that the EPUB file exists
	if _, err := os.Stat(epubPath); os.IsNotExist(err) {
		t.Fatalf("EPUB file does not exist")
	}

	// Open the EPUB file and check its contents
	zipReader, err := zip.OpenReader(epubPath)
	if err != nil {
		t.Fatalf("Failed to open EPUB: %v", err)
	}
	defer zipReader.Close()

	// Check for required EPUB files
	requiredFiles := []string{
		"mimetype",
		"META-INF/container.xml",
		"OEBPS/content.opf",
		"OEBPS/toc.ncx",
	}

	for _, requiredFile := range requiredFiles {
		found := false
		for _, file := range zipReader.File {
			if file.Name == requiredFile {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required file not found in EPUB: %s", requiredFile)
		}
	}

	// Check that the mimetype file is the first file in the archive
	if len(zipReader.File) > 0 && zipReader.File[0].Name != "mimetype" {
		t.Errorf("mimetype file is not the first file in the EPUB")
	}

	// Check that the mimetype file is stored uncompressed
	if len(zipReader.File) > 0 && zipReader.File[0].Method != zip.Store {
		t.Errorf("mimetype file is not stored uncompressed")
	}

	// Check that all images are included
	imageCount := 0
	for _, file := range zipReader.File {
		if strings.HasPrefix(file.Name, "OEBPS/images/") {
			imageCount++
		}
	}

	// Only count image files (not the text file)
	expectedImageCount := 0
	for _, image := range testImages {
		if isImageFile(image.name) {
			expectedImageCount++
		}
	}

	if imageCount != expectedImageCount {
		t.Errorf("Expected %d images in EPUB, got %d", expectedImageCount, imageCount)
	}

	// Check that all HTML pages are included
	pageCount := 0
	for _, file := range zipReader.File {
		if strings.HasPrefix(file.Name, "OEBPS/pages/") {
			pageCount++
		}
	}

	if pageCount != expectedImageCount {
		t.Errorf("Expected %d pages in EPUB, got %d", expectedImageCount, pageCount)
	}
}

// TestConvertFile tests the ConvertFile function
func TestConvertFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "epub_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test CBZ file
	testCBZPath := filepath.Join(tempDir, "test.cbz")
	testImages := []struct{ name, content string }{
		{"image1.jpg", "test image 1 content"},
		{"image2.png", "test image 2 content"},
	}
	createTestCBZ(t, testCBZPath, testImages)

	// Convert the CBZ to EPUB
	epubPath := filepath.Join(tempDir, "test.epub")
	err = ConvertFile(testCBZPath, epubPath)
	if err != nil {
		t.Fatalf("ConvertFile failed: %v", err)
	}

	// Check that the EPUB file exists
	if _, err := os.Stat(epubPath); os.IsNotExist(err) {
		t.Fatalf("EPUB file does not exist")
	}

	// Test with a non-existent file
	err = ConvertFile(filepath.Join(tempDir, "nonexistent.cbz"), filepath.Join(tempDir, "nonexistent.epub"))
	if err == nil {
		t.Errorf("ConvertFile should fail with non-existent file")
	}
}
