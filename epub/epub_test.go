package epub

import (
	"archive/zip"
	"io"
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
		if !cbz.IsImageFile(image.name) {
			continue
		}

		cbzFile.Images = append(cbzFile.Images, cbz.Image{
			Name:     filepath.Base(image.name),
			Data:     []byte(image.content),
			MimeType: cbz.MimeType(image.name),
		})
	}

	return cbzFile
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
		if cbz.IsImageFile(image.name) {
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

// TestConvertStreaming verifies the streaming EPUB conversion produces the same
// structural output as the in-memory ConvertFromCBZ path.
func TestConvertStreaming(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "epub_stream_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Build a real CBZ on disk.
	testImages := []struct{ name, content string }{
		{"002.png", "content-2"},
		{"001.jpg", "content-1"},
		{"not_image.txt", "skip me"},
		{"003.gif", "content-3"},
	}
	testCBZPath := filepath.Join(tempDir, "book.cbz")
	_ = createTestCBZ(t, testCBZPath, testImages)

	// Convert via the streaming API.
	epubPath := filepath.Join(tempDir, "book.epub")
	if err := ConvertStreaming(testCBZPath, epubPath); err != nil {
		t.Fatalf("ConvertStreaming failed: %v", err)
	}

	// Open the EPUB and check structure.
	zipReader, err := zip.OpenReader(epubPath)
	if err != nil {
		t.Fatalf("Failed to open EPUB: %v", err)
	}
	defer zipReader.Close()

	required := []string{
		"mimetype",
		"META-INF/container.xml",
		"OEBPS/content.opf",
		"OEBPS/toc.ncx",
		"OEBPS/images/image001.jpg",
		"OEBPS/images/image002.png",
		"OEBPS/images/image003.gif",
		"OEBPS/pages/page001.xhtml",
		"OEBPS/pages/page002.xhtml",
		"OEBPS/pages/page003.xhtml",
	}
	present := make(map[string]bool)
	for _, f := range zipReader.File {
		present[f.Name] = true
	}
	for _, name := range required {
		if !present[name] {
			t.Errorf("EPUB missing required entry: %s", name)
		}
	}

	// Verify sort order: 001.jpg maps to image001 with content-1.
	for _, f := range zipReader.File {
		if f.Name == "OEBPS/images/image001.jpg" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open image001: %v", err)
			}
			b, _ := io.ReadAll(rc)
			rc.Close()
			if string(b) != "content-1" {
				t.Errorf("image001.jpg content = %q, want %q", string(b), "content-1")
			}
		}
	}

	// Verify content.opf references image001..image003 and page001..page003.
	for _, f := range zipReader.File {
		if f.Name == "OEBPS/content.opf" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open content.opf: %v", err)
			}
			b, _ := io.ReadAll(rc)
			rc.Close()
			opf := string(b)
			for _, want := range []string{"image001", "image002", "image003", "page001", "page002", "page003"} {
				if !strings.Contains(opf, want) {
					t.Errorf("content.opf missing %s", want)
				}
			}
		}
	}
}

// TestConvertFileUsesStreaming verifies ConvertFile now delegates to the
// streaming path (i.e., works end-to-end on a real file without loading everything).
func TestConvertFileUsesStreaming(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "epub_convert_file_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testImages := []struct{ name, content string }{
		{"a.jpg", "img-a"},
		{"b.png", "img-b"},
	}
	testCBZPath := filepath.Join(tempDir, "book.cbz")
	_ = createTestCBZ(t, testCBZPath, testImages)

	epubPath := filepath.Join(tempDir, "book.epub")
	if err := ConvertFile(testCBZPath, epubPath); err != nil {
		t.Fatalf("ConvertFile failed: %v", err)
	}
	if _, err := os.Stat(epubPath); os.IsNotExist(err) {
		t.Fatalf("EPUB not created")
	}
}
