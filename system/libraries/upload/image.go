package upload

import (
	"bytes"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

// ImageConfig holds image processing configuration
type ImageConfig struct {
	// Source image path
	Source string

	// Destination path (empty = overwrite source)
	Destination string

	// Width for resize
	Width int

	// Height for resize
	Height int

	// Quality for JPEG (1-100)
	Quality int

	// MaintainAspectRatio
	MaintainAspectRatio bool

	// CreateThumbnail creates a thumbnail with prefix
	CreateThumbnail bool
	ThumbnailPrefix string
	ThumbnailWidth  int
	ThumbnailHeight int
}

// ImageProcessor handles image manipulation
type ImageProcessor struct {
	config ImageConfig
	errors []string
}

// Image errors
var (
	ErrInvalidImage     = errors.New("invalid image file")
	ErrUnsupportedType  = errors.New("unsupported image type")
	ErrProcessFailed    = errors.New("failed to process image")
)

// NewImageProcessor creates a new image processor
func NewImageProcessor(config ImageConfig) *ImageProcessor {
	if config.Quality == 0 {
		config.Quality = 85
	}
	if config.ThumbnailPrefix == "" {
		config.ThumbnailPrefix = "thumb_"
	}
	return &ImageProcessor{
		config: config,
		errors: make([]string, 0),
	}
}

// Resize resizes an image
func (ip *ImageProcessor) Resize() error {
	// Open source image
	srcFile, err := os.Open(ip.config.Source)
	if err != nil {
		return ErrInvalidImage
	}
	defer srcFile.Close()

	// Decode image
	img, format, err := image.Decode(srcFile)
	if err != nil {
		return ErrInvalidImage
	}

	// Get original bounds
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// Calculate new dimensions
	newWidth, newHeight := ip.calculateDimensions(srcWidth, srcHeight)

	// Create new image
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Resize using high-quality scaling
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	// Determine destination
	dest := ip.config.Destination
	if dest == "" {
		dest = ip.config.Source
	}

	// Save the image
	if err := ip.saveImage(dst, dest, format); err != nil {
		return err
	}

	// Create thumbnail if requested
	if ip.config.CreateThumbnail {
		if err := ip.createThumbnail(img, bounds, format); err != nil {
			return err
		}
	}

	return nil
}

// Crop crops an image
func (ip *ImageProcessor) Crop(x, y, width, height int) error {
	// Open source image
	srcFile, err := os.Open(ip.config.Source)
	if err != nil {
		return ErrInvalidImage
	}
	defer srcFile.Close()

	// Decode image
	img, format, err := image.Decode(srcFile)
	if err != nil {
		return ErrInvalidImage
	}

	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Validate crop dimensions
	if x < 0 || y < 0 || x+width > imgWidth || y+height > imgHeight {
		return ErrInvalidImage
	}

	// Create cropped image
	cropRect := image.Rect(0, 0, width, height)
	dst := image.NewRGBA(cropRect)

	// Copy cropped region
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			dst.Set(dx, dy, img.At(x+dx, y+dy))
		}
	}

	// Determine destination
	dest := ip.config.Destination
	if dest == "" {
		dest = ip.config.Source
	}

	// Save the image
	return ip.saveImage(dst, dest, format)
}

// Fit resizes image to fit within bounds while maintaining aspect ratio
func (ip *ImageProcessor) Fit() error {
	return ip.Resize()
}

// Fill resizes and crops image to fill exact dimensions
func (ip *ImageProcessor) Fill() error {
	// Open source image
	srcFile, err := os.Open(ip.config.Source)
	if err != nil {
		return ErrInvalidImage
	}
	defer srcFile.Close()

	// Decode image
	img, format, err := image.Decode(srcFile)
	if err != nil {
		return ErrInvalidImage
	}

	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	targetWidth := ip.config.Width
	targetHeight := ip.config.Height

	if targetWidth == 0 || targetHeight == 0 {
		return ErrInvalidImage
	}

	// Calculate scale to fill
	scaleX := float64(targetWidth) / float64(srcWidth)
	scaleY := float64(targetHeight) / float64(srcHeight)
	scale := scaleX
	if scaleY > scaleX {
		scale = scaleY
	}

	// Calculate new dimensions after scaling
	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)

	// First resize
	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)

	// Then crop center
	cropX := (newWidth - targetWidth) / 2
	cropY := (newHeight - targetHeight) / 2

	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	for dy := 0; dy < targetHeight; dy++ {
		for dx := 0; dx < targetWidth; dx++ {
			dst.Set(dx, dy, resized.At(cropX+dx, cropY+dy))
		}
	}

	// Determine destination
	dest := ip.config.Destination
	if dest == "" {
		dest = ip.config.Source
	}

	// Save the image
	return ip.saveImage(dst, dest, format)
}

// Rotate rotates an image by 90, 180, or 270 degrees
func (ip *ImageProcessor) Rotate(degrees int) error {
	// Open source image
	srcFile, err := os.Open(ip.config.Source)
	if err != nil {
		return ErrInvalidImage
	}
	defer srcFile.Close()

	// Decode image
	img, format, err := image.Decode(srcFile)
	if err != nil {
		return ErrInvalidImage
	}

	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	var dst *image.RGBA

	switch degrees {
	case 90:
		dst = image.NewRGBA(image.Rect(0, 0, srcHeight, srcWidth))
		for y := 0; y < srcHeight; y++ {
			for x := 0; x < srcWidth; x++ {
				dst.Set(srcHeight-1-y, x, img.At(x, y))
			}
		}
	case 180:
		dst = image.NewRGBA(image.Rect(0, 0, srcWidth, srcHeight))
		for y := 0; y < srcHeight; y++ {
			for x := 0; x < srcWidth; x++ {
				dst.Set(srcWidth-1-x, srcHeight-1-y, img.At(x, y))
			}
		}
	case 270:
		dst = image.NewRGBA(image.Rect(0, 0, srcHeight, srcWidth))
		for y := 0; y < srcHeight; y++ {
			for x := 0; x < srcWidth; x++ {
				dst.Set(y, srcWidth-1-x, img.At(x, y))
			}
		}
	default:
		return ErrInvalidImage
	}

	// Determine destination
	dest := ip.config.Destination
	if dest == "" {
		dest = ip.config.Source
	}

	// Save the image
	return ip.saveImage(dst, dest, format)
}

// calculateDimensions calculates new dimensions while maintaining aspect ratio
func (ip *ImageProcessor) calculateDimensions(srcWidth, srcHeight int) (int, int) {
	newWidth := ip.config.Width
	newHeight := ip.config.Height

	if newWidth == 0 && newHeight == 0 {
		return srcWidth, srcHeight
	}

	if ip.config.MaintainAspectRatio || (newWidth == 0 || newHeight == 0) {
		aspectRatio := float64(srcWidth) / float64(srcHeight)

		if newWidth == 0 {
			newWidth = int(float64(newHeight) * aspectRatio)
		} else if newHeight == 0 {
			newHeight = int(float64(newWidth) / aspectRatio)
		} else {
			// Fit within bounds
			targetRatio := float64(newWidth) / float64(newHeight)
			if aspectRatio > targetRatio {
				newHeight = int(float64(newWidth) / aspectRatio)
			} else {
				newWidth = int(float64(newHeight) * aspectRatio)
			}
		}
	}

	return newWidth, newHeight
}

// createThumbnail creates a thumbnail from source image
func (ip *ImageProcessor) createThumbnail(img image.Image, bounds image.Rectangle, format string) error {
	thumbWidth := ip.config.ThumbnailWidth
	thumbHeight := ip.config.ThumbnailHeight

	if thumbWidth == 0 {
		thumbWidth = 150
	}
	if thumbHeight == 0 {
		thumbHeight = 150
	}

	// Calculate thumbnail dimensions maintaining aspect ratio
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()
	aspectRatio := float64(srcWidth) / float64(srcHeight)

	var newWidth, newHeight int
	if aspectRatio > 1 {
		newWidth = thumbWidth
		newHeight = int(float64(thumbWidth) / aspectRatio)
	} else {
		newHeight = thumbHeight
		newWidth = int(float64(thumbHeight) * aspectRatio)
	}

	// Create thumbnail
	thumb := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), img, bounds, draw.Over, nil)

	// Create thumbnail path
	dir := filepath.Dir(ip.config.Source)
	filename := filepath.Base(ip.config.Source)
	thumbPath := filepath.Join(dir, ip.config.ThumbnailPrefix+filename)

	// Save thumbnail
	return ip.saveImage(thumb, thumbPath, format)
}

// saveImage saves an image to file
func (ip *ImageProcessor) saveImage(img image.Image, path, format string) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if dir != "" {
		os.MkdirAll(dir, 0755)
	}

	// Create file
	out, err := os.Create(path)
	if err != nil {
		return ErrProcessFailed
	}
	defer out.Close()

	// Encode based on format
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return jpeg.Encode(out, img, &jpeg.Options{Quality: ip.config.Quality})
	case "png":
		return png.Encode(out, img)
	case "gif":
		return gif.Encode(out, img, nil)
	default:
		return png.Encode(out, img)
	}
}

// DeleteImage deletes an image file
func DeleteImage(path string) error {
	return os.Remove(path)
}

// DeleteWithThumbnail deletes an image and its thumbnail
func DeleteWithThumbnail(path, thumbPrefix string) error {
	// Delete main image
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Delete thumbnail
	if thumbPrefix != "" {
		dir := filepath.Dir(path)
		filename := filepath.Base(path)
		thumbPath := filepath.Join(dir, thumbPrefix+filename)
		os.Remove(thumbPath) // Ignore error if thumb doesn't exist
	}

	return nil
}

// ReadImage reads image from reader and returns image.Image
func ReadImage(r io.Reader) (image.Image, string, error) {
	return image.Decode(r)
}

// WriteImage writes image to writer
func WriteImage(w io.Writer, img image.Image, format string, quality int) error {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	case "png":
		return png.Encode(w, img)
	case "gif":
		return gif.Encode(w, img, nil)
	default:
		return png.Encode(w, img)
	}
}

// ImageToBytes converts image to byte slice
func ImageToBytes(img image.Image, format string, quality int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := WriteImage(buf, img, format, quality)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// BytesToImage converts byte slice to image
func BytesToImage(data []byte) (image.Image, string, error) {
	return image.Decode(bytes.NewReader(data))
}