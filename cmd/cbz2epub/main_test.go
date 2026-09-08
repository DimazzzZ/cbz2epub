package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseFlags tests the parseFlags function
func TestParseFlags(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		args           []string
		expectedConfig Config
	}{
		{
			name: "merge command",
			args: []string{"-merge", "file1.cbz", "file2.cbz"},
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
			args: []string{"-merge", "-output", "merged.cbz", "file1.cbz", "file2.cbz"},
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
			args: []string{"-convert", "file.cbz"},
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
			args: []string{"-convert", "-output", "file.epub", "file.cbz"},
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
			args: []string{"-convert", "-verbose", "file.cbz"},
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
			args: []string{"-convert", "-recursive", "directory"},
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
			args: []string{},
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
			// Call parseFlags with the args directly; no global state involved.
			config, err := parseFlags(tc.args)
			if err != nil {
				t.Fatalf("parseFlags returned unexpected error: %v", err)
			}

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
// TestWalkCBZFiles tests the walkCBZFiles function
func TestWalkCBZFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "walk_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a nested directory structure:
	//   tempDir/
	//     a.cbz
	//     b.txt
	//     sub/
	//       c.cbz
	//       deep/
	//         d.cbz
	for _, name := range []string{"a.cbz", "b.txt"} {
		if err := os.WriteFile(filepath.Join(tempDir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	subDir := filepath.Join(tempDir, "sub")
	deepDir := filepath.Join(subDir, "deep")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{filepath.Join(subDir, "c.cbz"), filepath.Join(deepDir, "d.cbz")} {
		if err := os.WriteFile(name, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		inputs    []string
		recursive bool
		wantCount int
		wantErr   bool
	}{
		{
			name:      "single cbz file",
			inputs:    []string{filepath.Join(tempDir, "a.cbz")},
			recursive: false,
			wantCount: 1,
		},
		{
			name:      "non-cbz file is skipped",
			inputs:    []string{filepath.Join(tempDir, "b.txt")},
			recursive: false,
			wantCount: 0,
		},
		{
			name:      "directory without recursive is skipped",
			inputs:    []string{subDir},
			recursive: false,
			wantCount: 0,
		},
		{
			name:      "directory with recursive collects nested files",
			inputs:    []string{tempDir},
			recursive: true,
			wantCount: 3, // a.cbz, sub/c.cbz, sub/deep/d.cbz
		},
		{
			name:      "mixed files and dirs",
			inputs:    []string{filepath.Join(tempDir, "a.cbz"), subDir},
			recursive: true,
			wantCount: 3, // a.cbz, sub/c.cbz, sub/deep/d.cbz
		},
		{
			name:      "nonexistent path returns error",
			inputs:    []string{filepath.Join(tempDir, "nope.cbz")},
			recursive: false,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := walkCBZFiles(tc.inputs, tc.recursive, false)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tc.wantCount {
				t.Errorf("got %d files %v, want %d", len(got), got, tc.wantCount)
			}
		})
	}
}

// writeTestCBZ creates a minimal valid CBZ (zip with one image) at path.
func writeTestCBZ(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create cbz: %v", err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	w, err := zw.Create("image.jpg")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	if _, err := w.Write([]byte("fake image data")); err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
}

// TestExecute drives the real entry point end-to-end by passing args directly,
// exercising flag parsing + command dispatch without any global flag state.
func TestExecute(t *testing.T) {
	tempDir := t.TempDir()

	cbzA := filepath.Join(tempDir, "a.cbz")
	cbzB := filepath.Join(tempDir, "b.cbz")
	writeTestCBZ(t, cbzA)
	writeTestCBZ(t, cbzB)

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantOutput string // if non-empty, assert this file exists after run
	}{
		{
			name:    "no command prints usage and succeeds",
			args:    []string{},
			wantErr: false,
		},
		{
			name:       "convert single file with explicit output",
			args:       []string{"-convert", "-output", filepath.Join(tempDir, "out.epub"), cbzA},
			wantErr:    false,
			wantOutput: filepath.Join(tempDir, "out.epub"),
		},
		{
			name:       "merge two files with explicit output",
			args:       []string{"-merge", "-output", filepath.Join(tempDir, "merged.cbz"), cbzA, cbzB},
			wantErr:    false,
			wantOutput: filepath.Join(tempDir, "merged.cbz"),
		},
		{
			name:    "convert with no input files errors",
			args:    []string{"-convert"},
			wantErr: true,
		},
		{
			name:    "merge with no input files errors",
			args:    []string{"-merge"},
			wantErr: true,
		},
		{
			name:    "convert nonexistent file errors",
			args:    []string{"-convert", filepath.Join(tempDir, "nope.cbz")},
			wantErr: true,
		},
		{
			name:    "unknown flag errors",
			args:    []string{"-bogus"},
			wantErr: true,
		},
		{
			name:    "version flag prints version and succeeds",
			args:    []string{"-version"},
			wantErr: false,
		},
		{
			name:       "convert single file with default output name",
			args:       []string{"-convert", cbzA},
			wantErr:    false,
			wantOutput: strings.TrimSuffix(cbzA, ".cbz") + ".epub",
		},
		{
			name:    "convert recursive on directory with cbz files",
			args:    []string{"-convert", "-recursive", tempDir},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := execute(tc.args)
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantOutput != "" {
				if _, statErr := os.Stat(tc.wantOutput); statErr != nil {
					t.Errorf("expected output %s to exist: %v", tc.wantOutput, statErr)
				}
			}
		})
	}
}
