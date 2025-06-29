package cbz2epub

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cbz2epub/cbz"
	"cbz2epub/epub"
)

// Config holds the application configuration
type Config struct {
	Merge      bool
	Convert    bool
	OutputFile string
	Verbose    bool
	Recursive  bool
	InputFiles []string
}

// Execute runs the application
func Execute() error {
	// Set up logging
	log.SetPrefix("[CBZ2EPUB] ")
	log.SetFlags(log.LstdFlags)

	// Parse command line flags
	config := parseFlags()

	// Process commands
	if config.Merge {
		return handleMergeCommand(config)
	} else if config.Convert {
		return handleConvertCommand(config)
	} else {
		printUsage()
		return nil
	}
}

// parseFlags parses command line flags and returns a Config
func parseFlags() Config {
	// Define command line flags
	mergeCmd := flag.Bool("merge", false, "Merge multiple CBZ files into one")
	convertCmd := flag.Bool("convert", false, "Convert CBZ to EPUB")
	outputFile := flag.String("output", "", "Output file name")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	recursive := flag.Bool("recursive", false, "Process directories recursively")

	flag.Parse()

	// Get input files
	inputFiles := flag.Args()

	// If no input files specified, check if we should process current directory
	if len(inputFiles) == 0 && *recursive {
		// Get all CBZ files in current directory
		files, err := filepath.Glob("*.cbz")
		if err == nil && len(files) > 0 {
			inputFiles = files
		}
	}

	return Config{
		Merge:      *mergeCmd,
		Convert:    *convertCmd,
		OutputFile: *outputFile,
		Verbose:    *verbose,
		Recursive:  *recursive,
		InputFiles: inputFiles,
	}
}

// handleMergeCommand handles the merge command
func handleMergeCommand(config Config) error {
	if len(config.InputFiles) == 0 {
		log.Println("No input files specified")
		printUsage()
		return fmt.Errorf("no input files specified")
	}

	// Sort input files by name to ensure proper order
	sort.Strings(config.InputFiles)

	// Set default output file if not specified
	outputFile := config.OutputFile
	if outputFile == "" {
		outputFile = "merged.cbz"
	}

	if config.Verbose {
		log.Printf("Merging %d files into %s\n", len(config.InputFiles), outputFile)
	}

	// Merge files
	err := cbz.MergeFiles(config.InputFiles, outputFile)
	if err != nil {
		log.Printf("Error merging CBZ files: %v", err)
		return err
	}

	log.Printf("Successfully merged %d CBZ files into %s\n", len(config.InputFiles), outputFile)
	return nil
}

// handleConvertCommand handles the convert command
func handleConvertCommand(config Config) error {
	if len(config.InputFiles) == 0 {
		log.Println("No input files specified")
		printUsage()
		return fmt.Errorf("no input files specified")
	}

	var conversionError error

	// Process each input file
	for _, inputFile := range config.InputFiles {
		// Check if it's a directory
		fileInfo, err := os.Stat(inputFile)
		if err != nil {
			log.Printf("Error accessing %s: %v\n", inputFile, err)
			conversionError = err
			continue
		}

		if fileInfo.IsDir() {
			if config.Recursive {
				if err := processDirectory(inputFile, config); err != nil {
					conversionError = err
				}
			} else {
				log.Printf("Skipping directory %s (use -recursive to process directories)\n", inputFile)
			}
			continue
		}

		// Process single file
		if !strings.HasSuffix(strings.ToLower(inputFile), ".cbz") {
			log.Printf("Skipping non-CBZ file: %s\n", inputFile)
			continue
		}

		// Set output file name
		outputFile := config.OutputFile
		if outputFile == "" || len(config.InputFiles) > 1 {
			outputFile = strings.TrimSuffix(inputFile, ".cbz") + ".epub"
		}

		if config.Verbose {
			log.Printf("Converting %s to %s\n", inputFile, outputFile)
		}

		// Convert file
		err = epub.ConvertFile(inputFile, outputFile)
		if err != nil {
			log.Printf("Error converting %s: %v\n", inputFile, err)
			conversionError = err
			continue
		}

		log.Printf("Successfully converted %s to %s\n", inputFile, outputFile)
	}

	return conversionError
}

// processDirectory processes all CBZ files in a directory
func processDirectory(dirPath string, config Config) error {
	if config.Verbose {
		log.Printf("Processing directory: %s\n", dirPath)
	}

	var processingError error

	// Find all CBZ files in the directory
	files, err := filepath.Glob(filepath.Join(dirPath, "*.cbz"))
	if err != nil {
		log.Printf("Error finding CBZ files in %s: %v\n", dirPath, err)
		return err
	}

	if len(files) == 0 {
		log.Printf("No CBZ files found in %s\n", dirPath)
		return nil
	}

	// Process each file
	for _, file := range files {
		outputFile := strings.TrimSuffix(file, ".cbz") + ".epub"

		if config.Verbose {
			log.Printf("Converting %s to %s\n", file, outputFile)
		}

		err := epub.ConvertFile(file, outputFile)
		if err != nil {
			log.Printf("Error converting %s: %v\n", file, err)
			processingError = err
			continue
		}

		log.Printf("Successfully converted %s to %s\n", file, outputFile)
	}

	// If recursive, process subdirectories
	if config.Recursive {
		subdirs, err := os.ReadDir(dirPath)
		if err != nil {
			log.Printf("Error reading subdirectories in %s: %v\n", dirPath, err)
			return err
		}

		for _, subdir := range subdirs {
			if subdir.IsDir() {
				if err := processDirectory(filepath.Join(dirPath, subdir.Name()), config); err != nil && processingError == nil {
					processingError = err
				}
			}
		}
	}

	return processingError
}

// printUsage prints the usage information
func printUsage() {
	fmt.Println("CBZ2EPUB - A tool for merging CBZ files and converting them to EPUB")
	fmt.Println("\nUsage:")
	fmt.Println("  cbz2epub -merge [-output filename.cbz] file1.cbz file2.cbz ...")
	fmt.Println("  cbz2epub -convert [-output filename.epub] file.cbz")
	fmt.Println("  cbz2epub -convert -recursive [directory]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
}
