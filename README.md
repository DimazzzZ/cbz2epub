# CBZ2EPUB

CBZ2EPUB is a command-line tool for working with comic book archives. It allows you to merge multiple CBZ (Comic Book ZIP) files into a single file and convert CBZ files to EPUB format for e-readers.

## Features

- Merge multiple CBZ files into one, with proper renaming to avoid conflicts
- Convert CBZ files to EPUB format
- Process files in bulk with recursive directory scanning
- Simple command-line interface

## Installation

### Pre-built Binaries

Pre-built binaries for various platforms are available on the [Releases](https://github.com/DimazzzZ/cbz2epub/releases) page.

### Building from Source

#### Prerequisites

- Go 1.18 or later

#### Build Instructions

##### Linux/macOS

1. Clone the repository:
   ```bash
   git clone https://github.com/DimazzzZ/cbz2epub.git
   cd cbz2epub
   ```

2. Build the application:
   ```bash
   go build -o cbz2epub
   ```

3. (Optional) Install the application to your PATH:
   ```bash
   sudo mv cbz2epub /usr/local/bin/
   ```

##### Windows

1. Clone the repository:
   ```cmd
   git clone https://github.com/DimazzzZ/cbz2epub.git
   cd cbz2epub
   ```

2. Build the application:
   ```cmd
   go build -o cbz2epub.exe
   ```

3. (Optional) Add the directory to your PATH or move the executable to a directory in your PATH.

## Usage

### Basic Commands

```
CBZ2EPUB - A tool for merging CBZ files and converting them to EPUB

Usage:
  cbz2epub -merge [-output filename.cbz] file1.cbz file2.cbz ...
  cbz2epub -convert [-output filename.epub] file.cbz
  cbz2epub -convert -recursive [directory]

Options:
  -convert
        Convert CBZ to EPUB
  -merge
        Merge multiple CBZ files into one
  -output string
        Output file name
  -recursive
        Process directories recursively
  -verbose
        Enable verbose output
```

### Examples

#### Merging CBZ Files

Merge multiple CBZ files into a single file:

```bash
cbz2epub -merge -output merged.cbz chapter1.cbz chapter2.cbz chapter3.cbz
```

or just all files in the dir:

```bash
cbz2epub -merge -output merged.cbz chapter*.cbz
```

If no output file is specified, the default name "merged.cbz" will be used:

```bash
cbz2epub -merge chapter1.cbz chapter2.cbz chapter3.cbz
```

#### Converting CBZ to EPUB

Convert a single CBZ file to EPUB:

```bash
cbz2epub -convert -output mycomic.epub comic.cbz
```

If no output file is specified, the output filename will be derived from the input filename:

```bash
cbz2epub -convert comic.cbz
# Creates comic.epub
```

#### Bulk Conversion

Convert all CBZ files in the current directory:

```bash
cbz2epub -convert -recursive .
```

Convert all CBZ files in a specific directory and its subdirectories:

```bash
cbz2epub -convert -recursive /path/to/comics
```

#### Verbose Output

Add the `-verbose` flag to get more detailed output:

```bash
cbz2epub -convert -verbose -recursive /path/to/comics
```

## License

This project is licensed under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
