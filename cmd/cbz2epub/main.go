package main

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

// version is the application version. It is overridden at build time via
// -ldflags "-X main.version=<tag>" in the release workflow; it defaults to
// "dev" for source builds.
var version = "dev"

// main is the entry point for the cbz2epub application.
func main() {
	if err := execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Config holds the application configuration
type Config struct {
	Merge      bool
	Convert    bool
	OutputFile string
	Verbose    bool
	Recursive  bool
	InputFiles []string
	Version    bool
}

// execute runs the application with the given command-line arguments.
// It accepts args (typically os.Args[1:]) so the entry point is testable
// without touching global flag state.
func execute(args []string) error {
	// Set up logging
	log.SetPrefix("[CBZ2EPUB] ")
	log.SetFlags(log.LstdFlags)

	// Parse command line flags
	config, err := parseFlags(args)
	if err != nil {
		return err
	}

	// Process commands
	if config.Version {
		fmt.Fprintf(os.Stdout, "cbz2epub %s\n", version)
		return nil
	} else if config.Merge {
		return handleMergeCommand(config)
	} else if config.Convert {
		return handleConvertCommand(config)
	} else {
		printUsage()
		return nil
	}
}

// parseFlags parses command line flags from the given args and returns a Config.
// It uses a local FlagSet (ContinueOnError) instead of the global flag.CommandLine
// so it can be called repeatedly and tested without global-state gymnastics.
func parseFlags(args []string) (Config, error) {
	// Define command line flags on a local flag set.
	fs := flag.NewFlagSet("cbz2epub", flag.ContinueOnError)
	mergeCmd := fs.Bool("merge", false, "Merge multiple CBZ files into one")
	convertCmd := fs.Bool("convert", false, "Convert CBZ to EPUB")
	outputFile := fs.String("output", "", "Output file name")
	verbose := fs.Bool("verbose", false, "Enable verbose output")
	recursive := fs.Bool("recursive", false, "Process directories recursively")
	showVersion := fs.Bool("version", false, "Print version and exit")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	// Get input files
	inputFiles := fs.Args()

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
		Version:    *showVersion,
	}, nil
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

	// Collect all .cbz file paths from the inputs.
	cbzFiles, err := walkCBZFiles(config.InputFiles, config.Recursive, config.Verbose)
	if err != nil {
		return err
	}

	// Convert each collected file.
	var conversionError error
	// The -output flag is only honored when the user passed exactly one input
	// argument and it resolved to a single file (not a directory walk).
	useExplicitOutput := config.OutputFile != "" &&
		len(config.InputFiles) == 1 &&
		len(cbzFiles) == 1 &&
		cbzFiles[0] == config.InputFiles[0]
	for _, inputFile := range cbzFiles {
		outputFile := config.OutputFile
		if !useExplicitOutput {
			outputFile = strings.TrimSuffix(inputFile, ".cbz") + ".epub"
		}

		if config.Verbose {
			log.Printf("Converting %s to %s\n", inputFile, outputFile)
		}

		if err := epub.ConvertFile(inputFile, outputFile); err != nil {
			log.Printf("Error converting %s: %v\n", inputFile, err)
			conversionError = err
			continue
		}

		log.Printf("Successfully converted %s to %s\n", inputFile, outputFile)
	}

	return conversionError
}

// walkCBZFiles resolves a list of file/directory arguments into a flat list of
// .cbz file paths.  When recursive is true, directories are walked; otherwise
// they are skipped with a log message.
func walkCBZFiles(inputs []string, recursive, verbose bool) ([]string, error) {
	var cbzFiles []string

	for _, input := range inputs {
		info, err := os.Stat(input)
		if err != nil {
			log.Printf("Error accessing %s: %v\n", input, err)
			return nil, err
		}

		if !info.IsDir() {
			if strings.HasSuffix(strings.ToLower(input), ".cbz") {
				cbzFiles = append(cbzFiles, input)
			} else {
				log.Printf("Skipping non-CBZ file: %s\n", input)
			}
			continue
		}

		if !recursive {
			log.Printf("Skipping directory %s (use -recursive to process directories)\n", input)
			continue
		}

		if verbose {
			log.Printf("Processing directory: %s\n", input)
		}

		collected, err := collectCBZInDir(input, verbose)
		if err != nil {
			return nil, err
		}
		cbzFiles = append(cbzFiles, collected...)
	}

	return cbzFiles, nil
}

// collectCBZInDir recursively collects .cbz file paths under dirPath.
func collectCBZInDir(dirPath string, verbose bool) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dirPath, "*.cbz"))
	if err != nil {
		log.Printf("Error finding CBZ files in %s: %v\n", dirPath, err)
		return nil, err
	}

	cbzFiles := append([]string{}, matches...)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Printf("Error reading subdirectories in %s: %v\n", dirPath, err)
		return cbzFiles, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			sub, err := collectCBZInDir(filepath.Join(dirPath, entry.Name()), verbose)
			if err != nil {
				return cbzFiles, err
			}
			cbzFiles = append(cbzFiles, sub...)
		}
	}

	return cbzFiles, nil
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
