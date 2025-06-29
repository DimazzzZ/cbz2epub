package cbz2epub

import (
	"archive/zip"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// TestParseFlags tests the parseFlags function
func TestParseFlags(t *testing.T) {
	// Save original command line arguments and flags
	oldArgs := os.Args
	oldFlagCommandLine := flag.CommandLine
	defer func() {
		// Restore original command line arguments and flags
		os.Args = oldArgs
		flag.CommandLine = oldFlagCommandLine
	}()

	// Test cases
	testCases := []struct {
		name           string
		args           []string
		expectedConfig Config
	}{
		{
			name: "merge command",
			args: []string{"cbz2epub", "-merge", "file1.cbz", "file2.cbz"},
			expectedConfig: Config{
				Merge:      true,
				Convert:    false,
				OutputFile: "",
				Verbose:    false,
				Recursive:  false,
				InputFiles: []string{"file1.cbz", "file2.cbz"},
			},
		},
		{
			name: "merge command with output",
			args: []string{"cbz2epub", "-merge", "-output", "merged.cbz", "file1.cbz", "file2.cbz"},
			expectedConfig: Config{
				Merge:      true,
				Convert:    false,
				OutputFile: "merged.cbz",
				Verbose:    false,
				Recursive:  false,
				InputFiles: []string{"file1.cbz", "file2.cbz"},
			},
		},
		{
			name: "convert command",
			args: []string{"cbz2epub", "-convert", "file.cbz"},
			expectedConfig: Config{
				Merge:      false,
				Convert:    true,
				OutputFile: "",
				Verbose:    false,
				Recursive:  false,
				InputFiles: []string{"file.cbz"},
			},
		},
		{
			name: "convert command with output",
			args: []string{"cbz2epub", "-convert", "-output", "file.epub", "file.cbz"},
			expectedConfig: Config{
				Merge:      false,
				Convert:    true,
				OutputFile: "file.epub",
				Verbose:    false,
				Recursive:  false,
				InputFiles: []string{"file.cbz"},
			},
		},
		{
			name: "convert command with verbose",
			args: []string{"cbz2epub", "-convert", "-verbose", "file.cbz"},
			expectedConfig: Config{
				Merge:      false,
				Convert:    true,
				OutputFile: "",
				Verbose:    true,
				Recursive:  false,
				InputFiles: []string{"file.cbz"},
			},
		},
		{
			name: "convert command with recursive",
			args: []string{"cbz2epub", "-convert", "-recursive", "directory"},
			expectedConfig: Config{
				Merge:      false,
				Convert:    true,
				OutputFile: "",
				Verbose:    false,
				Recursive:  true,
				InputFiles: []string{"directory"},
			},
		},
		{
			name: "no command",
			args: []string{"cbz2epub"},
			expectedConfig: Config{
				Merge:      false,
				Convert:    false,
				OutputFile: "",
				Verbose:    false,
				Recursive:  false,
				InputFiles: []string{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flags
			flag.CommandLine = flag.NewFlagSet(tc.args[0], flag.ExitOnError)
			// Set command line arguments
			os.Args = tc.args

			// Call parseFlags
			config := parseFlags()

			// Check results
			if config.Merge != tc.expectedConfig.Merge {
				t.Errorf("Expected Merge=%v, got %v", tc.expectedConfig.Merge, config.Merge)
			}
			if config.Convert != tc.expectedConfig.Convert {
				t.Errorf("Expected Convert=%v, got %v", tc.expectedConfig.Convert, config.Convert)
			}
			if config.OutputFile != tc.expectedConfig.OutputFile {
				t.Errorf("Expected OutputFile=%v, got %v", tc.expectedConfig.OutputFile, config.OutputFile)
			}
			if config.Verbose != tc.expectedConfig.Verbose {
				t.Errorf("Expected Verbose=%v, got %v", tc.expectedConfig.Verbose, config.Verbose)
			}
			if config.Recursive != tc.expectedConfig.Recursive {
				t.Errorf("Expected Recursive=%v, got %v", tc.expectedConfig.Recursive, config.Recursive)
			}
			if len(config.InputFiles) != len(tc.expectedConfig.InputFiles) {
				t.Errorf("Expected %d input files, got %d", len(tc.expectedConfig.InputFiles), len(config.InputFiles))
			} else {
				for i, file := range tc.expectedConfig.InputFiles {
					if config.InputFiles[i] != file {
						t.Errorf("Expected InputFiles[%d]=%v, got %v", i, file, config.InputFiles[i])
					}
				}
			}
		})
	}
}

// TestHandleMergeCommand tests the handleMergeCommand function
func TestHandleMergeCommand(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "cbz2epub_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test CBZ files (valid zip files)
	testFile1 := filepath.Join(tempDir, "test1.cbz")
	testFile2 := filepath.Join(tempDir, "test2.cbz")

	// Create valid zip files
	for _, file := range []string{testFile1, testFile2} {
		zipFile, err := os.Create(file)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		zipWriter := zip.NewWriter(zipFile)

		// Add a dummy file to make it a valid zip
		writer, err := zipWriter.Create("dummy.txt")
		if err != nil {
			t.Fatalf("Failed to create file in test zip: %v", err)
		}

		_, err = writer.Write([]byte("dummy content"))
		if err != nil {
			t.Fatalf("Failed to write data in test zip: %v", err)
		}

		// Add an image file to make it a valid CBZ
		imageWriter, err := zipWriter.Create("image.jpg")
		if err != nil {
			t.Fatalf("Failed to create image in test zip: %v", err)
		}

		_, err = imageWriter.Write([]byte("fake image data"))
		if err != nil {
			t.Fatalf("Failed to write image data in test zip: %v", err)
		}

		zipWriter.Close()
		zipFile.Close()
	}

	// Test cases
	testCases := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "merge with output",
			config: Config{
				Merge:      true,
				OutputFile: filepath.Join(tempDir, "merged.cbz"),
				InputFiles: []string{testFile1, testFile2},
			},
			expectError: false,
		},
		{
			name: "merge without output",
			config: Config{
				Merge:      true,
				OutputFile: "",
				InputFiles: []string{testFile1, testFile2},
			},
			expectError: false,
		},
		{
			name: "merge with no input files",
			config: Config{
				Merge:      true,
				OutputFile: filepath.Join(tempDir, "merged.cbz"),
				InputFiles: []string{},
			},
			expectError: true,
		},
		{
			name: "merge with non-existent input file",
			config: Config{
				Merge:      true,
				OutputFile: filepath.Join(tempDir, "merged.cbz"),
				InputFiles: []string{filepath.Join(tempDir, "nonexistent.cbz")},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := handleMergeCommand(tc.config)
			if tc.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Check if output file exists when no error is expected
			if !tc.expectError && tc.config.OutputFile != "" {
				if _, err := os.Stat(tc.config.OutputFile); os.IsNotExist(err) {
					t.Errorf("Output file does not exist: %s", tc.config.OutputFile)
				}
			}
		})
	}
}

// TestHandleConvertCommand tests the handleConvertCommand function
func TestHandleConvertCommand(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "cbz2epub_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test CBZ file (valid zip file)
	testFile := filepath.Join(tempDir, "test.cbz")
	zipFile, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	zipWriter := zip.NewWriter(zipFile)

	// Add a dummy file to make it a valid zip
	writer, err := zipWriter.Create("dummy.txt")
	if err != nil {
		t.Fatalf("Failed to create file in test zip: %v", err)
	}

	_, err = writer.Write([]byte("dummy content"))
	if err != nil {
		t.Fatalf("Failed to write data in test zip: %v", err)
	}

	// Add an image file to make it a valid CBZ
	imageWriter, err := zipWriter.Create("image.jpg")
	if err != nil {
		t.Fatalf("Failed to create image in test zip: %v", err)
	}

	_, err = imageWriter.Write([]byte("fake image data"))
	if err != nil {
		t.Fatalf("Failed to write image data in test zip: %v", err)
	}

	zipWriter.Close()
	zipFile.Close()

	// Create test directory
	testDir := filepath.Join(tempDir, "testdir")
	err = os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test cases
	testCases := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "convert with output",
			config: Config{
				Convert:    true,
				OutputFile: filepath.Join(tempDir, "test.epub"),
				InputFiles: []string{testFile},
			},
			expectError: false, // Should succeed with valid CBZ file
		},
		{
			name: "convert without output",
			config: Config{
				Convert:    true,
				OutputFile: "",
				InputFiles: []string{testFile},
			},
			expectError: false, // Should succeed with valid CBZ file
		},
		{
			name: "convert with no input files",
			config: Config{
				Convert:    true,
				OutputFile: filepath.Join(tempDir, "test.epub"),
				InputFiles: []string{},
			},
			expectError: true,
		},
		{
			name: "convert with non-existent input file",
			config: Config{
				Convert:    true,
				OutputFile: filepath.Join(tempDir, "test.epub"),
				InputFiles: []string{filepath.Join(tempDir, "nonexistent.cbz")},
			},
			expectError: true,
		},
		{
			name: "convert with directory",
			config: Config{
				Convert:    true,
				Recursive:  false,
				OutputFile: "",
				InputFiles: []string{testDir},
			},
			expectError: false, // Should not error, just skip the directory
		},
		{
			name: "convert with directory and recursive",
			config: Config{
				Convert:    true,
				Recursive:  true,
				OutputFile: "",
				InputFiles: []string{testDir},
			},
			expectError: false, // Should not error, just process the directory
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := handleConvertCommand(tc.config)
			if tc.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// TestExecute is a placeholder test for the Execute function
// Testing the actual Execute function is complex due to global flag state
// and would require significant mocking. Instead, we test the individual
// components (parseFlags, handleMergeCommand, handleConvertCommand) separately.
func TestExecute(t *testing.T) {
	// This is a placeholder test to ensure coverage
	// The actual functionality is tested in other tests
	t.Skip("Skipping TestExecute as it requires complex mocking")
}
