package upload

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds upload configuration
type Config struct {
	// UploadPath is the destination directory (relative to project root or absolute)
	UploadPath string

	// AllowedTypes is a pipe-separated list of allowed file extensions (e.g., "gif|jpg|jpeg|png|pdf")
	AllowedTypes string

	// MaxSize is the maximum file size in KB (0 = no limit)
	MaxSize int64

	// FileName determines how the file is named: "original", "random", "timestamp", or custom name
	FileName string

	// Overwrite whether to overwrite existing files
	Overwrite bool

	// CreateDirs whether to create directories if they don't exist
	CreateDirs bool

	// FileExt to force a specific extension (e.g., ".jpg")
	FileExt string

	// Image settings
	MaxWidth  int
	MaxHeight int
	MinWidth  int
	MinHeight int
}

// Result holds upload result information
type Result struct {
	FileName      string
	OriginalName  string
	FileType      string
	FilePath      string
	FileSize      int64
	FileExt       string
	IsImage       bool
	ImageWidth    int
	ImageHeight   int
	ImageType     string
}

// Upload handles file uploads
type Upload struct {
	config    Config
	file      *multipart.FileHeader
	result    *Result
	errors    []string
	mimeType  string
}

// Common errors
var (
	ErrNoFile         = errors.New("no file was uploaded")
	ErrFileTooBig     = errors.New("the uploaded file exceeds the maximum allowed size")
	ErrInvalidType    = errors.New("the filetype you are attempting to upload is not allowed")
	ErrNoUploadPath   = errors.New("the upload path does not appear to be valid")
	ErrInvalidFile    = errors.New("the uploaded file is not a valid file")
	ErrWriteFailed    = errors.New("unable to write the file to disk")
	ErrFileExists     = errors.New("a file with the same name already exists")
	ErrInvalidDim     = errors.New("the image dimensions are invalid")
)

// New creates a new Upload instance with default configuration
func New(config Config) *Upload {
	return &Upload{
		config: config,
		errors: make([]string, 0),
	}
}

// Do performs the file upload
func (u *Upload) Do(field string, r *http.Request) (*Result, error) {
	// Reset state
	u.errors = make([]string, 0)
	u.result = nil

	// Parse multipart form if needed
	if r.MultipartForm == nil {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return nil, ErrNoFile
		}
	}

	// Get file headers
	fileHeaders, ok := r.MultipartForm.File[field]
	if !ok || len(fileHeaders) == 0 {
		return nil, ErrNoFile
	}

	u.file = fileHeaders[0]

	// Validate upload path
	if err := u.validatePath(); err != nil {
		return nil, err
	}

	// Validate file
	if err := u.validateFile(); err != nil {
		return nil, err
	}

	// Read file content
	src, err := u.file.Open()
	if err != nil {
		return nil, ErrInvalidFile
	}
	defer src.Close()

	// Generate filename
	filename := u.generateFilename()

	// Create full path
	fullPath := filepath.Join(u.config.UploadPath, filename)

	// Check if file exists
	if !u.config.Overwrite {
		if _, err := os.Stat(fullPath); err == nil {
			return nil, ErrFileExists
		}
	}

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, ErrWriteFailed
	}
	defer dst.Close()

	// Copy file
	written, err := io.Copy(dst, src)
	if err != nil {
		return nil, ErrWriteFailed
	}

	// Build result
	ext := strings.ToLower(filepath.Ext(filename))
	u.result = &Result{
		FileName:     filename,
		OriginalName: u.file.Filename,
		FileType:     u.mimeType,
		FilePath:     fullPath,
		FileSize:     written,
		FileExt:      ext,
		IsImage:      u.isImage(),
	}

	// Get image dimensions if applicable
	if u.result.IsImage {
		u.result.ImageWidth, u.result.ImageHeight, u.result.ImageType = GetImageDimensions(fullPath)
	}

	return u.result, nil
}

// DoWithFile performs upload with multipart.FileHeader directly
func (u *Upload) DoWithFile(fileHeader *multipart.FileHeader) (*Result, error) {
	u.errors = make([]string, 0)
	u.result = nil
	u.file = fileHeader

	// Validate upload path
	if err := u.validatePath(); err != nil {
		return nil, err
	}

	// Validate file
	if err := u.validateFile(); err != nil {
		return nil, err
	}

	// Read file content
	src, err := u.file.Open()
	if err != nil {
		return nil, ErrInvalidFile
	}
	defer src.Close()

	// Generate filename
	filename := u.generateFilename()

	// Create full path
	fullPath := filepath.Join(u.config.UploadPath, filename)

	// Check if file exists
	if !u.config.Overwrite {
		if _, err := os.Stat(fullPath); err == nil {
			return nil, ErrFileExists
		}
	}

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, ErrWriteFailed
	}
	defer dst.Close()

	// Copy file
	written, err := io.Copy(dst, src)
	if err != nil {
		return nil, ErrWriteFailed
	}

	// Build result
	ext := strings.ToLower(filepath.Ext(filename))
	u.result = &Result{
		FileName:     filename,
		OriginalName: u.file.Filename,
		FileType:     u.mimeType,
		FilePath:     fullPath,
		FileSize:     written,
		FileExt:      ext,
		IsImage:      u.isImage(),
	}

	// Get image dimensions if applicable
	if u.result.IsImage {
		u.result.ImageWidth, u.result.ImageHeight, u.result.ImageType = GetImageDimensions(fullPath)
	}

	return u.result, nil
}

// validatePath validates and creates the upload directory
func (u *Upload) validatePath() error {
	if u.config.UploadPath == "" {
		return ErrNoUploadPath
	}

	// Create directory if needed
	if u.config.CreateDirs {
		if err := os.MkdirAll(u.config.UploadPath, 0755); err != nil {
			return ErrNoUploadPath
		}
	}

	// Check if directory exists
	info, err := os.Stat(u.config.UploadPath)
	if err != nil || !info.IsDir() {
		return ErrNoUploadPath
	}

	return nil
}

// validateFile validates the uploaded file
func (u *Upload) validateFile() error {
	// Check file size
	if u.config.MaxSize > 0 && u.file.Size > u.config.MaxSize*1024 {
		return ErrFileTooBig
	}

	// Detect MIME type
	u.mimeType = GetMimeType(u.file)
	if u.mimeType == "" {
		u.mimeType = "application/octet-stream"
	}

	// Check allowed types
	if u.config.AllowedTypes != "" {
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(u.file.Filename)), ".")
		allowed := strings.Split(strings.ToLower(u.config.AllowedTypes), "|")
		if !contains(allowed, ext) {
			return ErrInvalidType
		}
	}

	return nil
}

// generateFilename generates the final filename
func (u *Upload) generateFilename() string {
	ext := filepath.Ext(u.file.Filename)

	// Force extension if specified
	if u.config.FileExt != "" {
		ext = u.config.FileExt
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
	}

	// Get base name without extension
	baseName := strings.TrimSuffix(u.file.Filename, filepath.Ext(u.file.Filename))

	switch u.config.FileName {
	case "random":
		return randomString(32) + ext
	case "timestamp":
		return fmt.Sprintf("%d_%s", time.Now().Unix(), sanitizeFilename(baseName)) + ext
	case "original", "":
		return sanitizeFilename(u.file.Filename)
	default:
		// Custom filename
		return u.config.FileName + ext
	}
}

// isImage checks if the file is an image
func (u *Upload) isImage() bool {
	imageTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp", "image/bmp"}
	return contains(imageTypes, u.mimeType)
}

// GetResult returns the upload result
func (u *Upload) GetResult() *Result {
	return u.result
}

// GetErrors returns all errors
func (u *Upload) GetErrors() []string {
	return u.errors
}

// Helper functions

// GetMimeType detects MIME type from file header
func GetMimeType(fileHeader *multipart.FileHeader) string {
	file, err := fileHeader.Open()
	if err != nil {
		return ""
	}
	defer file.Close()

	// Read first 512 bytes for detection
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return ""
	}

	// Reset file position
	file.Seek(0, 0)

	return http.DetectContentType(buffer)
}

// GetImageDimensions returns image dimensions
func GetImageDimensions(filepath string) (width, height int, imageType string) {
	file, err := os.Open(filepath)
	if err != nil {
		return 0, 0, ""
	}
	defer file.Close()

	// Read first 512 bytes
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return 0, 0, ""
	}

	// Detect image type and dimensions from magic bytes
	// JPEG
	if len(buffer) >= 2 && buffer[0] == 0xFF && buffer[1] == 0xD8 {
		return detectJPEGDimensions(filepath)
	}

	// PNG
	if len(buffer) >= 8 && buffer[0] == 0x89 && buffer[1] == 0x50 && buffer[2] == 0x4E && buffer[3] == 0x47 {
		return detectPNGDimensions(buffer), 3, "png"
	}

	// GIF
	if len(buffer) >= 6 && buffer[0] == 0x47 && buffer[1] == 0x49 && buffer[2] == 0x46 {
		width = int(buffer[6]) | int(buffer[7])<<8
		height = int(buffer[8]) | int(buffer[9])<<8
		return width, height, "gif"
	}

	return 0, 0, ""
}

// detectJPEGDimensions reads JPEG dimensions
func detectJPEGDimensions(filepath string) (int, int, string) {
	file, err := os.Open(filepath)
	if err != nil {
		return 0, 0, ""
	}
	defer file.Close()

	// Skip SOI marker
	buf := make([]byte, 2)
	file.Read(buf)

	for {
		marker := make([]byte, 2)
		if _, err := file.Read(marker); err != nil {
			break
		}

		// Check for SOF markers
		if marker[0] == 0xFF {
			switch marker[1] {
			case 0xC0, 0xC1, 0xC2, 0xC3, 0xC5, 0xC6, 0xC7, 0xC9, 0xCA, 0xCB, 0xCD, 0xCE, 0xCF:
				// Read segment length and dimensions
				seg := make([]byte, 7)
				file.Read(seg)
				height := int(seg[3])<<8 | int(seg[4])
				width := int(seg[5])<<8 | int(seg[6])
				return width, height, "jpeg"
			default:
				// Skip this segment
				segLen := make([]byte, 2)
				file.Read(segLen)
				length := int(segLen[0])<<8 | int(segLen[1])
				file.Seek(int64(length-2), 1)
			}
		}
	}

	return 0, 0, ""
}

// detectPNGDimensions reads PNG dimensions from buffer
func detectPNGDimensions(buffer []byte) int {
	if len(buffer) >= 24 {
		return int(buffer[16])<<24 | int(buffer[17])<<16 | int(buffer[18])<<8 | int(buffer[19])
	}
	return 0
}

// sanitizeFilename removes unsafe characters
func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	replace := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " ", "%"}
	for _, r := range replace {
		name = strings.ReplaceAll(name, r, "_")
	}
	return name
}

// randomString generates a random string
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}