package web

import (
	"archive/zip"
	"bytes"
	"coalesced-inifier/gow2"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func mainView(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]*template.Template)
	tmpl["index.html"] = template.Must(template.ParseFiles("templates/index.html", "templates/boilerplate.html"))

	err := tmpl["index.html"].ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func packView(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]*template.Template)
	tmpl["pack.html"] = template.Must(template.ParseFiles("templates/pack.html", "templates/boilerplate.html"))

	err := tmpl["pack.html"].ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func packAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(4 * 1024 * 1024)
	if err != nil {
		fmt.Println("Error retrieving file:", err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}

	// FormFile returns the first file for the provided form key
	file, h, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error retrieving file:", err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	zipReader, err := zip.NewReader(file, h.Size)
	if err != nil {
		http.Error(w, "Error reading uploaded file", http.StatusInternalServerError)
	}

	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	jobId := uuid.New().String()
	outputPath := filepath.Join(uploadDir, jobId)
	os.MkdirAll(outputPath, 0755)

	err = unzip(zipReader, outputPath)
	if err != nil {
		http.Error(w, "Error reading uploaded file", http.StatusInternalServerError)
	}

	fileList := make([]string, 0)
	err = recursiveFileList(outputPath, &fileList)
	if err != nil {
		http.Error(w, "Error reading uploaded file", http.StatusInternalServerError)
	}

	coalesced, err := gow2.Pack(fileList, outputPath, "..\\")

	buffer := new(bytes.Buffer)
	err = zipCoalesced(coalesced, buffer)
	if err != nil {
		http.Error(w, "Error reading uploaded file", http.StatusInternalServerError)
	}

	header := w.Header()
	header.Set("Content-Disposition", "attachment; filename=\"Coalesced.zip\"")
	header.Set("Content-Type", "application/zip")

	_, err = w.Write(buffer.Bytes())
	if err != nil {
		http.Error(w, "Error reading uploaded file", http.StatusInternalServerError)
	}
}

func unpackView(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]*template.Template)
	tmpl["unpack.html"] = template.Must(template.ParseFiles("templates/unpack.html", "templates/boilerplate.html"))

	err := tmpl["unpack.html"].ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func unpackAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(4 * 1024 * 1024)
	if err != nil {
		fmt.Println("Error retrieving file:", err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}

	// FormFile returns the first file for the provided form key
	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error retrieving file:", err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	jobId := uuid.New().String()
	outputPath := filepath.Join(uploadDir, jobId)
	os.MkdirAll(outputPath, 0755)

	buf := make([]byte, 4*1024*1024)
	n, err := file.Read(buf)
	if err != nil {
		http.Error(w, "Error reading uploaded file", http.StatusInternalServerError)
	}

	// TODO parse game prefix from form
	// TODO sanitize paths
	err = gow2.Unpack(bytes.NewReader(buf[:n]), outputPath, "..\\")
	if err != nil {
		http.Error(w, "Error reading uploaded file", http.StatusInternalServerError)
	}

	err, zipFile := zipFolder(outputPath)
	if err != nil {
		http.Error(w, "Error creating destination file", http.StatusInternalServerError)
		return
	}

	header := w.Header()
	header.Set("Content-Disposition", "attachment; filename=\""+jobId+".zip\"")
	header.Set("Content-Type", "application/zip")

	_, err = w.Write(zipFile.Bytes())
	if err != nil {
		http.Error(w, "Error reading uploaded file", http.StatusInternalServerError)
	}
}

func Hook(port int16) error {
	http.HandleFunc("/", mainView)
	http.HandleFunc("/pack", packView)
	http.HandleFunc("/pack-handle", packAction)
	http.HandleFunc("/unpack", unpackView)
	http.HandleFunc("/unpack-handle", unpackAction)
	fmt.Printf("Server started on :%d\n", port)
	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
}

func zipCoalesced(coalesced *bytes.Buffer, buffer *bytes.Buffer) error {
	zipWriter := zip.NewWriter(buffer)
	defer zipWriter.Close()

	hash := md5.Sum(coalesced.Bytes())

	writer, err := zipWriter.Create("Coalesced_int.bin")
	if err != nil {
		return err
	}

	_, err = writer.Write(coalesced.Bytes())
	if err != nil {
		return err
	}

	writer, err = zipWriter.Create("tocinfo.txt")
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "File length: %d\n", len(coalesced.Bytes()))
	fmt.Fprintf(writer, "File hash: %s\n\n", hex.EncodeToString(hash[:]))

	return err
}

func zipFolder(sourceFolder string) (error, *bytes.Buffer) {
	buffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buffer)
	defer zipWriter.Close()

	err := filepath.Walk(sourceFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, sourceFolder)
		if len(header.Name) > 0 && header.Name[0] == '/' {
			header.Name = header.Name[1:]
		}

		if info.IsDir() {
			header.Name += "/"
			_, err = zipWriter.CreateHeader(header)
			return err
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return err, buffer
}

const (
	// Set reasonable limits
	maxFileSize     = 1024 * 1024 * 100 // 100MB per file
	maxTotalSize    = 1024 * 1024 * 500 // 500MB total extraction
	maxPathLength   = 256
	maxFiles        = 1000
	maxSymlinkDepth = 8
)

func unzip(r *zip.Reader, destDir string) error {
	// Ensure destination exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	destDirAbs, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute destination path: %w", err)
	}

	// Check number of files
	if len(r.File) > maxFiles {
		return fmt.Errorf("too many files in archive: limit is %d, got %d", maxFiles, len(r.File))
	}

	// Track total size extracted
	var totalExtracted int64

	// Process each file
	for _, f := range r.File {
		// Check for path manipulation (zipslip)
		destPath := filepath.Join(destDirAbs, f.Name)
		if !strings.HasPrefix(destPath, destDirAbs+string(os.PathSeparator)) {
			return fmt.Errorf("zip slip detected: %s", f.Name)
		}

		// Check path length
		if len(destPath) > maxPathLength {
			return fmt.Errorf("path too long: %s (%d characters)", f.Name, len(destPath))
		}

		// Check file size
		if f.UncompressedSize64 > uint64(maxFileSize) {
			return fmt.Errorf("maximum file size exceeded: %s (%d bytes)", f.Name, f.UncompressedSize64)
		}

		// Update and check total extracted size
		totalExtracted += int64(f.UncompressedSize64)
		if totalExtracted > maxTotalSize {
			return fmt.Errorf("maximum total extraction size exceeded: limit is %d bytes", maxTotalSize)
		}

		// Handle directories
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
			continue
		}

		// Ensure parent directory exists for files
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", f.Name, err)
		}

		// Create the file
		destFile, err := safeCreateFile(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", destPath, err)
		}

		// Open the file from the archive
		srcFile, err := f.Open()
		if err != nil {
			destFile.Close()
			return fmt.Errorf("failed to open file from archive %s: %w", f.Name, err)
		}

		// Copy with size limit enforcement
		copiedBytes, err := io.Copy(destFile, io.LimitReader(srcFile, maxFileSize))
		srcFile.Close()
		destFile.Close()

		if err != nil {
			return fmt.Errorf("failed to extract file %s: %w", f.Name, err)
		}

		// Verify size matches expected (detect truncation)
		if copiedBytes != int64(f.UncompressedSize64) {
			return fmt.Errorf("size mismatch for %s: expected %d, got %d",
				f.Name, f.UncompressedSize64, copiedBytes)
		}
	}

	return nil
}

func safeCreateFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

func recursiveFileList(current string, inputFiles *[]string) error {
	dir, err := os.ReadDir(current)
	if err != nil {
		return err
	}

	for i := 0; i < len(dir); i++ {
		entry := dir[i]
		entryPath := path.Join(current, entry.Name())
		if entry.IsDir() {
			err = recursiveFileList(entryPath, inputFiles)
		} else if isIniOrInt(entryPath) {
			*inputFiles = append(*inputFiles, entryPath)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func isIniOrInt(file string) bool {
	file = strings.ToLower(file)
	return strings.HasSuffix(file, ".int") || strings.HasSuffix(file, ".ini")
}
