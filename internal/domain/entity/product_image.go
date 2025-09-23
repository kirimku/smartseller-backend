package entity

import (
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ProductImage represents an image associated with a product or product variant
type ProductImage struct {
	// Primary identification
	ID        uuid.UUID  `json:"id" db:"id"`
	ProductID uuid.UUID  `json:"product_id" db:"product_id"`
	VariantID *uuid.UUID `json:"variant_id" db:"variant_id"` // NULL for product-level images

	// Image properties
	ImageURL  string  `json:"image_url" db:"image_url"`
	AltText   *string `json:"alt_text" db:"alt_text"`
	IsPrimary bool    `json:"is_primary" db:"is_primary"`
	SortOrder int     `json:"sort_order" db:"sort_order"`

	// File information
	FileName *string `json:"file_name" db:"file_name"`
	FileSize *int64  `json:"file_size" db:"file_size"` // Size in bytes
	MimeType *string `json:"mime_type" db:"mime_type"` // e.g., "image/jpeg"
	Width    *int    `json:"width" db:"width"`         // Image width in pixels
	Height   *int    `json:"height" db:"height"`       // Image height in pixels

	// External service metadata
	CloudinaryID  *string `json:"cloudinary_id" db:"cloudinary_id"`   // Cloudinary public ID
	CloudinaryURL *string `json:"cloudinary_url" db:"cloudinary_url"` // Cloudinary secure URL
	ThumbnailURL  *string `json:"thumbnail_url" db:"thumbnail_url"`   // Optimized thumbnail
	MediumURL     *string `json:"medium_url" db:"medium_url"`         // Medium size URL
	LargeURL      *string `json:"large_url" db:"large_url"`           // Large size URL

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Computed fields
	FileExtension string `json:"file_extension" db:"-"`
	FormattedSize string `json:"formatted_size" db:"-"`
	AspectRatio   string `json:"aspect_ratio" db:"-"`
	IsOptimized   bool   `json:"is_optimized" db:"-"`
	ImageType     string `json:"image_type" db:"-"` // "product" or "variant"
}

// ImageType constants
const (
	ImageTypeProduct = "product"
	ImageTypeVariant = "variant"
)

// Allowed image MIME types
var AllowedImageMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"image/bmp":  true,
	"image/tiff": true,
}

// NewProductImage creates a new product image
func NewProductImage(productID uuid.UUID, imageURL string) *ProductImage {
	return &ProductImage{
		ID:        uuid.New(),
		ProductID: productID,
		ImageURL:  imageURL,
		IsPrimary: false,
		SortOrder: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewProductVariantImage creates a new product variant image
func NewProductVariantImage(productID, variantID uuid.UUID, imageURL string) *ProductImage {
	return &ProductImage{
		ID:        uuid.New(),
		ProductID: productID,
		VariantID: &variantID,
		ImageURL:  imageURL,
		IsPrimary: false,
		SortOrder: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ValidateImageURL validates the image URL
func (pi *ProductImage) ValidateImageURL() error {
	if pi.ImageURL == "" {
		return fmt.Errorf("image URL is required")
	}

	if len(pi.ImageURL) > 2048 {
		return fmt.Errorf("image URL cannot exceed 2048 characters")
	}

	// Parse URL
	parsedURL, err := url.Parse(pi.ImageURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	// Check if URL has a valid image extension or mime type indication
	ext := strings.ToLower(filepath.Ext(parsedURL.Path))
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff"}

	isValidExt := false
	for _, validExt := range validExtensions {
		if ext == validExt {
			isValidExt = true
			break
		}
	}

	// If no valid extension, check for Cloudinary or other CDN patterns
	if !isValidExt {
		// Allow Cloudinary URLs without extensions
		if strings.Contains(parsedURL.Host, "cloudinary.com") ||
			strings.Contains(parsedURL.Host, "res.cloudinary.com") {
			isValidExt = true
		}
		// Allow other common CDN patterns
		if strings.Contains(parsedURL.Path, "/image/") ||
			strings.Contains(parsedURL.Path, "/img/") ||
			strings.Contains(parsedURL.Path, "/photo/") {
			isValidExt = true
		}
	}

	if !isValidExt {
		return fmt.Errorf("URL does not appear to be a valid image URL")
	}

	return nil
}

// ValidateAltText validates the alt text
func (pi *ProductImage) ValidateAltText() error {
	if pi.AltText != nil {
		altText := strings.TrimSpace(*pi.AltText)
		if altText == "" {
			pi.AltText = nil // Clear empty alt text
			return nil
		}

		if len(altText) > 500 {
			return fmt.Errorf("alt text cannot exceed 500 characters")
		}

		pi.AltText = &altText
	}

	return nil
}

// ValidateFileInfo validates file-related information
func (pi *ProductImage) ValidateFileInfo() error {
	// Validate file name
	if pi.FileName != nil {
		fileName := strings.TrimSpace(*pi.FileName)
		if fileName == "" {
			pi.FileName = nil
		} else {
			if len(fileName) > 255 {
				return fmt.Errorf("file name cannot exceed 255 characters")
			}
			pi.FileName = &fileName
		}
	}

	// Validate file size
	if pi.FileSize != nil {
		if *pi.FileSize < 0 {
			return fmt.Errorf("file size cannot be negative")
		}

		// Maximum file size: 50MB
		maxSize := int64(50 * 1024 * 1024)
		if *pi.FileSize > maxSize {
			return fmt.Errorf("file size cannot exceed 50MB")
		}
	}

	// Validate MIME type
	if pi.MimeType != nil {
		mimeType := strings.ToLower(strings.TrimSpace(*pi.MimeType))
		if mimeType == "" {
			pi.MimeType = nil
		} else {
			if !AllowedImageMimeTypes[mimeType] {
				return fmt.Errorf("unsupported MIME type: %s", mimeType)
			}
			pi.MimeType = &mimeType
		}
	}

	// Validate dimensions
	if pi.Width != nil {
		if *pi.Width <= 0 {
			return fmt.Errorf("image width must be positive")
		}
		if *pi.Width > 10000 {
			return fmt.Errorf("image width cannot exceed 10,000 pixels")
		}
	}

	if pi.Height != nil {
		if *pi.Height <= 0 {
			return fmt.Errorf("image height must be positive")
		}
		if *pi.Height > 10000 {
			return fmt.Errorf("image height cannot exceed 10,000 pixels")
		}
	}

	return nil
}

// ValidateCloudinaryInfo validates Cloudinary-specific fields
func (pi *ProductImage) ValidateCloudinaryInfo() error {
	// Validate Cloudinary ID
	if pi.CloudinaryID != nil {
		cloudinaryID := strings.TrimSpace(*pi.CloudinaryID)
		if cloudinaryID == "" {
			pi.CloudinaryID = nil
		} else {
			if len(cloudinaryID) > 255 {
				return fmt.Errorf("Cloudinary ID cannot exceed 255 characters")
			}
			pi.CloudinaryID = &cloudinaryID
		}
	}

	// Validate Cloudinary URLs
	urlFields := map[string]**string{
		"CloudinaryURL": &pi.CloudinaryURL,
		"ThumbnailURL":  &pi.ThumbnailURL,
		"MediumURL":     &pi.MediumURL,
		"LargeURL":      &pi.LargeURL,
	}

	for fieldName, fieldPtr := range urlFields {
		if *fieldPtr != nil {
			urlStr := strings.TrimSpace(**fieldPtr)
			if urlStr == "" {
				*fieldPtr = nil
				continue
			}

			if len(urlStr) > 2048 {
				return fmt.Errorf("%s cannot exceed 2048 characters", fieldName)
			}

			// Validate URL format
			if _, err := url.Parse(urlStr); err != nil {
				return fmt.Errorf("invalid %s format: %w", fieldName, err)
			}

			**fieldPtr = urlStr
		}
	}

	return nil
}

// Validate performs comprehensive validation of the product image
func (pi *ProductImage) Validate() error {
	// Validate image URL
	if err := pi.ValidateImageURL(); err != nil {
		return fmt.Errorf("image URL validation failed: %w", err)
	}

	// Validate alt text
	if err := pi.ValidateAltText(); err != nil {
		return fmt.Errorf("alt text validation failed: %w", err)
	}

	// Validate file info
	if err := pi.ValidateFileInfo(); err != nil {
		return fmt.Errorf("file info validation failed: %w", err)
	}

	// Validate Cloudinary info
	if err := pi.ValidateCloudinaryInfo(); err != nil {
		return fmt.Errorf("Cloudinary info validation failed: %w", err)
	}

	// Sort order should not be negative
	if pi.SortOrder < 0 {
		return fmt.Errorf("sort order cannot be negative")
	}

	return nil
}

// SetAsPrimary sets this image as the primary image
func (pi *ProductImage) SetAsPrimary() {
	pi.IsPrimary = true
	pi.UpdatedAt = time.Now()
}

// UnsetAsPrimary removes primary status from this image
func (pi *ProductImage) UnsetAsPrimary() {
	pi.IsPrimary = false
	pi.UpdatedAt = time.Now()
}

// UpdateSortOrder updates the sort order
func (pi *ProductImage) UpdateSortOrder(sortOrder int) error {
	if sortOrder < 0 {
		return fmt.Errorf("sort order cannot be negative")
	}

	pi.SortOrder = sortOrder
	pi.UpdatedAt = time.Now()
	return nil
}

// SetAltText sets the alt text for the image
func (pi *ProductImage) SetAltText(altText string) error {
	trimmedText := strings.TrimSpace(altText)
	if trimmedText == "" {
		pi.AltText = nil
		pi.UpdatedAt = time.Now()
		return nil
	}

	if len(trimmedText) > 500 {
		return fmt.Errorf("alt text cannot exceed 500 characters")
	}

	pi.AltText = &trimmedText
	pi.UpdatedAt = time.Now()
	return nil
}

// UpdateFileInfo updates file information
func (pi *ProductImage) UpdateFileInfo(fileName *string, fileSize *int64, mimeType *string, width, height *int) error {
	// Create temporary copy for validation
	temp := *pi
	temp.FileName = fileName
	temp.FileSize = fileSize
	temp.MimeType = mimeType
	temp.Width = width
	temp.Height = height

	// Validate the changes
	if err := temp.ValidateFileInfo(); err != nil {
		return err
	}

	// Apply changes if validation passes
	pi.FileName = temp.FileName
	pi.FileSize = temp.FileSize
	pi.MimeType = temp.MimeType
	pi.Width = temp.Width
	pi.Height = temp.Height
	pi.UpdatedAt = time.Now()

	return nil
}

// UpdateCloudinaryInfo updates Cloudinary-specific information
func (pi *ProductImage) UpdateCloudinaryInfo(cloudinaryID, cloudinaryURL, thumbnailURL, mediumURL, largeURL *string) error {
	// Create temporary copy for validation
	temp := *pi
	temp.CloudinaryID = cloudinaryID
	temp.CloudinaryURL = cloudinaryURL
	temp.ThumbnailURL = thumbnailURL
	temp.MediumURL = mediumURL
	temp.LargeURL = largeURL

	// Validate the changes
	if err := temp.ValidateCloudinaryInfo(); err != nil {
		return err
	}

	// Apply changes if validation passes
	pi.CloudinaryID = temp.CloudinaryID
	pi.CloudinaryURL = temp.CloudinaryURL
	pi.ThumbnailURL = temp.ThumbnailURL
	pi.MediumURL = temp.MediumURL
	pi.LargeURL = temp.LargeURL
	pi.UpdatedAt = time.Now()

	return nil
}

// GetImageType returns whether this is a product or variant image
func (pi *ProductImage) GetImageType() string {
	if pi.VariantID != nil {
		return ImageTypeVariant
	}
	return ImageTypeProduct
}

// GetFileExtension extracts file extension from URL or filename
func (pi *ProductImage) GetFileExtension() string {
	if pi.FileName != nil {
		return strings.ToLower(filepath.Ext(*pi.FileName))
	}

	// Try to extract from URL
	parsedURL, err := url.Parse(pi.ImageURL)
	if err != nil {
		return ""
	}

	return strings.ToLower(filepath.Ext(parsedURL.Path))
}

// GetFormattedSize returns human-readable file size
func (pi *ProductImage) GetFormattedSize() string {
	if pi.FileSize == nil {
		return ""
	}

	size := *pi.FileSize
	units := []string{"B", "KB", "MB", "GB"}

	for i, unit := range units {
		if size < 1024 || i == len(units)-1 {
			if i == 0 {
				return fmt.Sprintf("%d %s", size, unit)
			}
			return fmt.Sprintf("%.1f %s", float64(size)/1024.0, unit)
		}
		size /= 1024
	}

	return ""
}

// GetAspectRatio calculates and returns aspect ratio
func (pi *ProductImage) GetAspectRatio() string {
	if pi.Width == nil || pi.Height == nil {
		return ""
	}

	width := *pi.Width
	height := *pi.Height

	// Calculate GCD for simplification
	gcd := func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}

	divisor := gcd(width, height)
	return fmt.Sprintf("%d:%d", width/divisor, height/divisor)
}

// IsOptimizedImage checks if the image has optimization URLs
func (pi *ProductImage) IsOptimizedImage() bool {
	return pi.ThumbnailURL != nil || pi.MediumURL != nil || pi.LargeURL != nil
}

// GetBestImageURL returns the best available image URL for a given size preference
func (pi *ProductImage) GetBestImageURL(sizePreference string) string {
	switch strings.ToLower(sizePreference) {
	case "thumbnail", "thumb", "small":
		if pi.ThumbnailURL != nil {
			return *pi.ThumbnailURL
		}
	case "medium", "med":
		if pi.MediumURL != nil {
			return *pi.MediumURL
		}
	case "large", "big":
		if pi.LargeURL != nil {
			return *pi.LargeURL
		}
	case "original", "full":
		if pi.CloudinaryURL != nil {
			return *pi.CloudinaryURL
		}
	}

	// Fallback to main image URL
	return pi.ImageURL
}

// DetectMimeTypeFromURL attempts to detect MIME type from URL extension
func (pi *ProductImage) DetectMimeTypeFromURL() string {
	ext := pi.GetFileExtension()
	if ext == "" {
		return ""
	}

	// Remove leading dot
	ext = strings.TrimPrefix(ext, ".")

	mimeType := mime.TypeByExtension("." + ext)
	if mimeType != "" {
		return strings.Split(mimeType, ";")[0] // Remove charset if present
	}

	// Manual mapping for common image types
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "bmp":
		return "image/bmp"
	case "tiff", "tif":
		return "image/tiff"
	default:
		return ""
	}
}

// ComputeFields calculates all computed fields
func (pi *ProductImage) ComputeFields() {
	pi.FileExtension = pi.GetFileExtension()
	pi.FormattedSize = pi.GetFormattedSize()
	pi.AspectRatio = pi.GetAspectRatio()
	pi.IsOptimized = pi.IsOptimizedImage()
	pi.ImageType = pi.GetImageType()

	// Auto-detect MIME type if not set
	if pi.MimeType == nil {
		detectedMime := pi.DetectMimeTypeFromURL()
		if detectedMime != "" {
			pi.MimeType = &detectedMime
		}
	}
}

// String returns a string representation of the product image
func (pi *ProductImage) String() string {
	imageType := pi.GetImageType()
	return fmt.Sprintf("ProductImage{ID: %s, ProductID: %s, Type: %s, IsPrimary: %t, URL: %s}",
		pi.ID.String(), pi.ProductID.String(), imageType, pi.IsPrimary, pi.ImageURL)
}
