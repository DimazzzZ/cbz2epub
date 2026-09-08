package epub

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cbz2epub/cbz"
)

// ConvertFromCBZ converts a CBZ file to EPUB format
func ConvertFromCBZ(cbzFile *cbz.File, outputFile string) error {
	// Create a new zip file for the EPUB
	zipFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add mimetype file (must be first and uncompressed)
	mimetypeWriter, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store, // No compression
	})
	if err != nil {
		return fmt.Errorf("failed to create mimetype file: %w", err)
	}
	_, err = mimetypeWriter.Write([]byte("application/epub+zip"))
	if err != nil {
		return fmt.Errorf("failed to write mimetype file: %w", err)
	}

	// Add META-INF/container.xml
	containerWriter, err := zipWriter.Create("META-INF/container.xml")
	if err != nil {
		return fmt.Errorf("failed to create container.xml: %w", err)
	}
	_, err = containerWriter.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))
	if err != nil {
		return fmt.Errorf("failed to write container.xml: %w", err)
	}

	// Create content.opf
	title := strings.TrimSuffix(filepath.Base(cbzFile.Name), ".cbz")
	date := time.Now().Format("2006-01-02")
	uuid := generateUUID()
	contentOPF := bytes.NewBufferString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="BookID" version="2.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:title>%s</dc:title>
    <dc:language>en</dc:language>
    <dc:identifier id="BookID">urn:uuid:%s</dc:identifier>
    <dc:date>%s</dc:date>
    <dc:creator>CBZ2EPUB Converter</dc:creator>
  </metadata>
  <manifest>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
`, title, uuid, date))

	// Add each image to the manifest
	for i, image := range cbzFile.Images {
		// Create a new name for the image to avoid conflicts
		ext := filepath.Ext(image.Name)
		newName := fmt.Sprintf("image%03d%s", i+1, ext)

		// Add to manifest
		contentOPF.WriteString(fmt.Sprintf(`    <item id="image%03d" href="images/%s" media-type="%s"/>
`, i+1, newName, image.MimeType))

		// Add to EPUB
		imageWriter, err := zipWriter.Create("OEBPS/images/" + newName)
		if err != nil {
			return fmt.Errorf("failed to create image file: %w", err)
		}
		_, err = imageWriter.Write(image.Data)
		if err != nil {
			return fmt.Errorf("failed to write image data: %w", err)
		}
	}

	// Create HTML pages for each image
	for i, image := range cbzFile.Images {
		// Create a new name for the image
		ext := filepath.Ext(image.Name)
		newName := fmt.Sprintf("image%03d%s", i+1, ext)

		// Create HTML page
		pageName := fmt.Sprintf("page%03d.xhtml", i+1)
		contentOPF.WriteString(fmt.Sprintf(`    <item id="page%03d" href="pages/%s" media-type="application/xhtml+xml"/>
`, i+1, pageName))

		// Add HTML page to EPUB
		pageWriter, err := zipWriter.Create("OEBPS/pages/" + pageName)
		if err != nil {
			return fmt.Errorf("failed to create page file: %w", err)
		}

		// Write HTML content
		_, err = pageWriter.Write([]byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Page %d</title>
  <style type="text/css">
    img { max-width: 100%%; max-height: 100%%; }
    body { margin: 0; padding: 0; text-align: center; }
  </style>
</head>
<body>
  <div>
    <img src="../images/%s" alt="Page %d" />
  </div>
</body>
</html>`, i+1, newName, i+1)))
		if err != nil {
			return fmt.Errorf("failed to write page content: %w", err)
		}
	}

	// Finish content.opf with spine
	contentOPF.WriteString(`  </manifest>
  <spine toc="ncx">
`)
	for i := range cbzFile.Images {
		contentOPF.WriteString(fmt.Sprintf(`    <itemref idref="page%03d"/>
`, i+1))
	}
	contentOPF.WriteString(`  </spine>
</package>`)

	// Add content.opf to EPUB
	contentWriter, err := zipWriter.Create("OEBPS/content.opf")
	if err != nil {
		return fmt.Errorf("failed to create content.opf: %w", err)
	}
	_, err = contentWriter.Write(contentOPF.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write content.opf: %w", err)
	}

	// Create toc.ncx
	tocNCX := bytes.NewBufferString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE ncx PUBLIC "-//NISO//DTD ncx 2005-1//EN" "http://www.daisy.org/z3986/2005/ncx-2005-1.dtd">
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
  <head>
    <meta name="dtb:uid" content="urn:uuid:%s"/>
    <meta name="dtb:depth" content="1"/>
    <meta name="dtb:totalPageCount" content="0"/>
    <meta name="dtb:maxPageNumber" content="0"/>
  </head>
  <docTitle>
    <text>%s</text>
  </docTitle>
  <navMap>
`, uuid, title))

	// Add each page to the navigation map
	for i := range cbzFile.Images {
		tocNCX.WriteString(fmt.Sprintf(`    <navPoint id="navpoint-%d" playOrder="%d">
      <navLabel>
        <text>Page %d</text>
      </navLabel>
      <content src="pages/page%03d.xhtml"/>
    </navPoint>
`, i+1, i+1, i+1, i+1))
	}
	tocNCX.WriteString(`  </navMap>
</ncx>`)

	// Add toc.ncx to EPUB
	tocWriter, err := zipWriter.Create("OEBPS/toc.ncx")
	if err != nil {
		return fmt.Errorf("failed to create toc.ncx: %w", err)
	}
	_, err = tocWriter.Write(tocNCX.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write toc.ncx: %w", err)
	}

	return nil
}

// generateUUID returns an RFC 4122 version 4 UUID string, falling back to a
// timestamp-based identifier if the system CSPRNG is unavailable.
func generateUUID() string {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}

	// Set version (4) and variant (RFC 4122).
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// ConvertStreaming converts a CBZ file to EPUB format, streaming images to
// avoid loading all into memory. This is the preferred method for large CBZ files.
func ConvertStreaming(inputFile, outputFile string) error {
	zipFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add mimetype file (must be first and uncompressed)
	mimetypeWriter, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	})
	if err != nil {
		return fmt.Errorf("failed to create mimetype file: %w", err)
	}
	_, err = mimetypeWriter.Write([]byte("application/epub+zip"))
	if err != nil {
		return fmt.Errorf("failed to write mimetype file: %w", err)
	}

	// Add META-INF/container.xml
	containerWriter, err := zipWriter.Create("META-INF/container.xml")
	if err != nil {
		return fmt.Errorf("failed to create container.xml: %w", err)
	}
	_, err = containerWriter.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))
	if err != nil {
		return fmt.Errorf("failed to write container.xml: %w", err)
	}

	title := strings.TrimSuffix(filepath.Base(inputFile), ".cbz")
	date := time.Now().Format("2006-01-02")
	uuid := generateUUID()

	// Collect image metadata while streaming images to the EPUB
	type imgMeta struct {
		index    int
		newName  string
		mimeType string
	}
	var images []imgMeta
	imageIndex := 1

	err = cbz.IterateImages(inputFile, func(name string, data io.Reader, mimeType string) error {
		ext := filepath.Ext(name)
		newName := fmt.Sprintf("image%03d%s", imageIndex, ext)

		// Write image bytes to EPUB
		imageWriter, err := zipWriter.Create("OEBPS/images/" + newName)
		if err != nil {
			return fmt.Errorf("failed to create image file: %w", err)
		}
		if _, err := io.Copy(imageWriter, data); err != nil {
			return fmt.Errorf("failed to write image data: %w", err)
		}

		// Write page XHTML
		pageName := fmt.Sprintf("page%03d.xhtml", imageIndex)
		pageWriter, err := zipWriter.Create("OEBPS/pages/" + pageName)
		if err != nil {
			return fmt.Errorf("failed to create page file: %w", err)
		}

		pageHTML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Page %d</title>
  <style type="text/css">
    img { max-width: 100%%; max-height: 100%%; }
    body { margin: 0; padding: 0; text-align: center; }
  </style>
</head>
<body>
  <div>
    <img src="../images/%s" alt="Page %d" />
  </div>
</body>
</html>`, imageIndex, newName, imageIndex)

		if _, err := pageWriter.Write([]byte(pageHTML)); err != nil {
			return fmt.Errorf("failed to write page content: %w", err)
		}

		images = append(images, imgMeta{index: imageIndex, newName: newName, mimeType: mimeType})
		imageIndex++
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to process CBZ images: %w", err)
	}

	// Build content.opf from accumulated metadata
	contentOPF := bytes.NewBufferString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="BookID" version="2.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:title>%s</dc:title>
    <dc:language>en</dc:language>
    <dc:identifier id="BookID">urn:uuid:%s</dc:identifier>
    <dc:date>%s</dc:date>
    <dc:creator>CBZ2EPUB Converter</dc:creator>
  </metadata>
  <manifest>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
`, title, uuid, date))

	for _, img := range images {
		contentOPF.WriteString(fmt.Sprintf(`    <item id="image%03d" href="images/%s" media-type="%s"/>
`, img.index, img.newName, img.mimeType))
	}

	for _, img := range images {
		contentOPF.WriteString(fmt.Sprintf(`    <item id="page%03d" href="pages/page%03d.xhtml" media-type="application/xhtml+xml"/>
`, img.index, img.index))
	}

	contentOPF.WriteString(`  </manifest>
  <spine toc="ncx">
`)
	for _, img := range images {
		contentOPF.WriteString(fmt.Sprintf(`    <itemref idref="page%03d"/>
`, img.index))
	}
	contentOPF.WriteString(`  </spine>
</package>`)

	contentWriter, err := zipWriter.Create("OEBPS/content.opf")
	if err != nil {
		return fmt.Errorf("failed to create content.opf: %w", err)
	}
	if _, err := contentWriter.Write(contentOPF.Bytes()); err != nil {
		return fmt.Errorf("failed to write content.opf: %w", err)
	}

	// Build and write toc.ncx
	tocNCX := bytes.NewBufferString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE ncx PUBLIC "-//NISO//DTD ncx 2005-1//EN" "http://www.daisy.org/z3986/2005/ncx-2005-1.dtd">
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
  <head>
    <meta name="dtb:uid" content="urn:uuid:%s"/>
    <meta name="dtb:depth" content="1"/>
    <meta name="dtb:totalPageCount" content="0"/>
    <meta name="dtb:maxPageNumber" content="0"/>
  </head>
  <docTitle>
    <text>%s</text>
  </docTitle>
  <navMap>
`, uuid, title))

	for _, img := range images {
		tocNCX.WriteString(fmt.Sprintf(`    <navPoint id="navpoint-%d" playOrder="%d">
      <navLabel>
        <text>Page %d</text>
      </navLabel>
      <content src="pages/page%03d.xhtml"/>
    </navPoint>
`, img.index, img.index, img.index, img.index))
	}
	tocNCX.WriteString(`  </navMap>
</ncx>`)

	tocWriter, err := zipWriter.Create("OEBPS/toc.ncx")
	if err != nil {
		return fmt.Errorf("failed to create toc.ncx: %w", err)
	}
	if _, err := tocWriter.Write(tocNCX.Bytes()); err != nil {
		return fmt.Errorf("failed to write toc.ncx: %w", err)
	}

	return nil
}

// ConvertFile converts a CBZ file to EPUB format using streaming.
func ConvertFile(inputFile, outputFile string) error {
	return ConvertStreaming(inputFile, outputFile)
}
