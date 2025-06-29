package cbz

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// TestIsImageFile tests the isImageFile function
func TestIsImageFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"image.jpg", true},
		{"image.jpeg", true},
		{"image.png", true},
		{"image.gif", true},
		{"image.webp", true},
		{"image.txt", false},
		{"image.pdf", false},
		{"image", false},
		{".jpg", true},
		{"image.JPG", true},
		{"image.PNG", true},
	}

	for _, test := range tests {
		result := isImageFile(test.filename)
		if result != test.expected {
			t.Errorf("isImageFile(%s) = %v, expected %v", test.filename, result, test.expected)
		}
	}
}

// TestGetMimeType tests the getMimeType function
func TestGetMimeType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"image.jpg", "image/jpeg"},
		{"image.jpeg", "image/jpeg"},
		{"image.png", "image/png"},
		{"image.gif", "image/gif"},
		{"image.webp", "image/webp"},
		{"image.txt", "application/octet-stream"},
		{"image.pdf", "application/octet-stream"},
		{"image", "application/octet-stream"},
		{".jpg", "image/jpeg"},
		{"image.JPG", "image/jpeg"},
		{"image.PNG", "image/png"},
	}

	for _, test := range tests {
		result := getMimeType(test.filename)
		if result != test.expected {
			t.Errorf("getMimeType(%s) = %v, expected %v", test.filename, result, test.expected)
		}
	}
}

// createTestCBZ creates a test CBZ file with the given images
func createTestCBZ(t *testing.T, filename string, images []struct{ name, content string }) {
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
}

// TestReadFile tests the ReadFile function
func TestReadFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "cbz_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test CBZ file
	testCBZ := filepath.Join(tempDir, "test.cbz")
	testImages := []struct{ name, content string }{
		{"image1.jpg", "test image 1 content"},
		{"image2.png", "test image 2 content"},
		{"subfolder/image3.gif", "test image 3 content"},
		{"not_an_image.txt", "this is not an image"},
	}
	createTestCBZ(t, testCBZ, testImages)

	// Test reading the CBZ file
	cbzFile, err := ReadFile(testCBZ)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// Check that the file name is correct
	if cbzFile.Name != testCBZ {
		t.Errorf("Expected file name %s, got %s", testCBZ, cbzFile.Name)
	}

	// Check that only image files were read (3 images, not the text file)
	if len(cbzFile.Images) != 3 {
		t.Errorf("Expected 3 images, got %d", len(cbzFile.Images))
	}

	// Check that the images are sorted by name
	if len(cbzFile.Images) >= 2 && cbzFile.Images[0].Name > cbzFile.Images[1].Name {
		t.Errorf("Images are not sorted by name")
	}

	// Check that the image data is correct
	for _, image := range cbzFile.Images {
		var found bool
		for _, testImage := range testImages {
			if filepath.Base(testImage.name) == image.Name {
				found = true
				if string(image.Data) != testImage.content {
					t.Errorf("Expected image content %s, got %s", testImage.content, string(image.Data))
				}
				break
			}
		}
		if !found {
			t.Errorf("Unexpected image found: %s", image.Name)
		}
	}
}

// TestMergeFiles tests the MergeFiles function
func TestMergeFiles(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "cbz_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create two test CBZ files
	testCBZ1 := filepath.Join(tempDir, "test1.cbz")
	testImages1 := []struct{ name, content string }{
		{"image1.jpg", "test image 1 content"},
		{"image2.png", "test image 2 content"},
	}
	createTestCBZ(t, testCBZ1, testImages1)

	testCBZ2 := filepath.Join(tempDir, "test2.cbz")
	testImages2 := []struct{ name, content string }{
		{"image1.jpg", "test image 3 content"}, // Same name as in test1.cbz
		{"image3.gif", "test image 4 content"},
	}
	createTestCBZ(t, testCBZ2, testImages2)

	// Merge the CBZ files
	mergedCBZ := filepath.Join(tempDir, "merged.cbz")
	err = MergeFiles([]string{testCBZ1, testCBZ2}, mergedCBZ)
	if err != nil {
		t.Fatalf("MergeFiles failed: %v", err)
	}

	// Check that the merged file exists
	if _, err := os.Stat(mergedCBZ); os.IsNotExist(err) {
		t.Fatalf("Merged file does not exist")
	}

	// Open the merged file and check its contents
	zipReader, err := zip.OpenReader(mergedCBZ)
	if err != nil {
		t.Fatalf("Failed to open merged CBZ: %v", err)
	}
	defer zipReader.Close()

	// Check that there are 4 files in the merged CBZ
	if len(zipReader.File) != 4 {
		t.Errorf("Expected 4 files in merged CBZ, got %d", len(zipReader.File))
	}

	// Check that the files have been renamed to avoid conflicts
	fileNames := make(map[string]bool)
	for _, file := range zipReader.File {
		fileNames[file.Name] = true
	}

	// Check that each file has a unique name
	if len(fileNames) != 4 {
		t.Errorf("Expected 4 unique file names, got %d", len(fileNames))
	}

	// Check that the file names follow the expected pattern (chapterXXX_YYY.ext)
	for fileName := range fileNames {
		if len(fileName) < 8 || fileName[:7] != "chapter" {
			t.Errorf("Unexpected file name format: %s", fileName)
		}
	}
}
