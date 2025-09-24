package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// ProductImageUseCase handles product image management business logic
type ProductImageUseCase struct {
	imageRepo   repository.ProductImageRepository
	productRepo repository.ProductRepository
	variantRepo repository.ProductVariantRepository
	logger      *slog.Logger
}

// NewProductImageUseCase creates a new product image use case
func NewProductImageUseCase(
	imageRepo repository.ProductImageRepository,
	productRepo repository.ProductRepository,
	variantRepo repository.ProductVariantRepository,
	logger *slog.Logger,
) *ProductImageUseCase {
	return &ProductImageUseCase{
		imageRepo:   imageRepo,
		productRepo: productRepo,
		variantRepo: variantRepo,
		logger:      logger,
	}
}

// CreateImageRequest represents a request to create a new product image
type CreateImageRequest struct {
	ProductID          uuid.UUID  `json:"product_id" validate:"required"`
	VariantID          *uuid.UUID `json:"variant_id,omitempty"`
	ImageURL           string     `json:"image_url" validate:"required,url"`
	CloudinaryURL      *string    `json:"cloudinary_url,omitempty"`
	CloudinaryPublicID *string    `json:"cloudinary_public_id,omitempty"`
	AltText            *string    `json:"alt_text,omitempty"`
	Width              *int       `json:"width,omitempty"`
	Height             *int       `json:"height,omitempty"`
	FileSize           *int64     `json:"file_size,omitempty"`
	MimeType           *string    `json:"mime_type,omitempty"`
	IsPrimary          *bool      `json:"is_primary,omitempty"`
	SortOrder          *int       `json:"sort_order,omitempty"`
}

// UpdateImageRequest represents a request to update an image
type UpdateImageRequest struct {
	ImageURL           *string `json:"image_url,omitempty"`
	CloudinaryURL      *string `json:"cloudinary_url,omitempty"`
	CloudinaryPublicID *string `json:"cloudinary_public_id,omitempty"`
	AltText            *string `json:"alt_text,omitempty"`
	Width              *int    `json:"width,omitempty"`
	Height             *int    `json:"height,omitempty"`
	FileSize           *int64  `json:"file_size,omitempty"`
	MimeType           *string `json:"mime_type,omitempty"`
	IsPrimary          *bool   `json:"is_primary,omitempty"`
	SortOrder          *int    `json:"sort_order,omitempty"`
}

// ReorderImagesRequest represents a request to reorder images
type ReorderImagesRequest struct {
	ProductID   uuid.UUID               `json:"product_id" validate:"required"`
	VariantID   *uuid.UUID              `json:"variant_id,omitempty"`
	ImageOrders []repository.ImageOrder `json:"image_orders" validate:"required,min=1"`
}

// SetPrimaryImageRequest represents a request to set primary image
type SetPrimaryImageRequest struct {
	ProductID uuid.UUID  `json:"product_id" validate:"required"`
	VariantID *uuid.UUID `json:"variant_id,omitempty"`
	ImageID   uuid.UUID  `json:"image_id" validate:"required"`
}

// BulkImageUploadRequest represents a request to upload multiple images
type BulkImageUploadRequest struct {
	ProductID uuid.UUID            `json:"product_id" validate:"required"`
	VariantID *uuid.UUID           `json:"variant_id,omitempty"`
	Images    []CreateImageRequest `json:"images" validate:"required,min=1,max=10"`
}

// ImageValidationOptions contains validation options for images
type ImageValidationOptions struct {
	MaxFileSize      int64    `json:"max_file_size"`   // bytes
	MaxImageCount    int      `json:"max_image_count"` // per product/variant
	AllowedMimeTypes []string `json:"allowed_mime_types"`
	MinWidth         int      `json:"min_width"`  // pixels
	MaxWidth         int      `json:"max_width"`  // pixels
	MinHeight        int      `json:"min_height"` // pixels
	MaxHeight        int      `json:"max_height"` // pixels
	RequireAltText   bool     `json:"require_alt_text"`
}

// DefaultImageValidationOptions returns default validation options
func DefaultImageValidationOptions() ImageValidationOptions {
	return ImageValidationOptions{
		MaxFileSize:      10 * 1024 * 1024, // 10MB
		MaxImageCount:    10,
		AllowedMimeTypes: []string{"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp"},
		MinWidth:         100,
		MaxWidth:         4000,
		MinHeight:        100,
		MaxHeight:        4000,
		RequireAltText:   false,
	}
}

// CreateImage creates a new product image
func (uc *ProductImageUseCase) CreateImage(ctx context.Context, req CreateImageRequest) (*entity.ProductImage, error) {
	uc.logger.Info("Creating product image",
		slog.String("product_id", req.ProductID.String()),
		slog.Any("variant_id", req.VariantID),
		slog.String("image_url", req.ImageURL))

	// Validate request
	if err := uc.validateCreateImageRequest(ctx, req); err != nil {
		uc.logger.Error("Image creation validation failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create image entity
	var image *entity.ProductImage
	if req.VariantID != nil {
		image = entity.NewProductVariantImage(req.ProductID, *req.VariantID, req.ImageURL)
	} else {
		image = entity.NewProductImage(req.ProductID, req.ImageURL)
	}

	// Apply optional fields
	uc.applyImageFields(image, req)

	// Set sort order if not provided
	if image.SortOrder == 0 {
		nextOrder, err := uc.imageRepo.GetNextSortOrder(ctx, req.ProductID, req.VariantID)
		if err != nil {
			uc.logger.Error("Failed to get next sort order", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to get next sort order: %w", err)
		}
		image.SortOrder = nextOrder
	}

	// Validate image constraints
	if err := uc.validateImageConstraints(ctx, image); err != nil {
		uc.logger.Error("Image constraints validation failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("image constraints validation failed: %w", err)
	}

	// Handle primary image logic
	if image.IsPrimary {
		if err := uc.handlePrimaryImageAssignment(ctx, image); err != nil {
			uc.logger.Error("Failed to handle primary image assignment", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to handle primary image assignment: %w", err)
		}
	}

	// Create image
	if err := uc.imageRepo.Create(ctx, image); err != nil {
		uc.logger.Error("Failed to create image", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create image: %w", err)
	}

	uc.logger.Info("Product image created successfully", slog.String("image_id", image.ID.String()))
	return image, nil
}

// GetImageByID retrieves an image by ID
func (uc *ProductImageUseCase) GetImageByID(ctx context.Context, imageID uuid.UUID) (*entity.ProductImage, error) {
	uc.logger.Info("Getting image by ID", slog.String("image_id", imageID.String()))

	if imageID == uuid.Nil {
		return nil, fmt.Errorf("image ID cannot be empty")
	}

	include := &repository.ProductImageInclude{
		Product: true,
		Variant: true,
	}

	image, err := uc.imageRepo.GetByID(ctx, imageID, include)
	if err != nil {
		uc.logger.Error("Failed to get image by ID",
			slog.String("image_id", imageID.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return image, nil
}

// UpdateImage updates an existing image
func (uc *ProductImageUseCase) UpdateImage(ctx context.Context, imageID uuid.UUID, req UpdateImageRequest) (*entity.ProductImage, error) {
	uc.logger.Info("Updating image", slog.String("image_id", imageID.String()))

	// Get existing image
	image, err := uc.GetImageByID(ctx, imageID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	uc.applyImageUpdates(image, req)

	// Validate constraints after updates
	if err := uc.validateImageConstraints(ctx, image); err != nil {
		uc.logger.Error("Image constraints validation failed after update", slog.String("error", err.Error()))
		return nil, fmt.Errorf("image constraints validation failed: %w", err)
	}

	// Handle primary image logic if changed
	if req.IsPrimary != nil && *req.IsPrimary && !image.IsPrimary {
		if err := uc.handlePrimaryImageAssignment(ctx, image); err != nil {
			uc.logger.Error("Failed to handle primary image assignment", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to handle primary image assignment: %w", err)
		}
	}

	// Update image
	if err := uc.imageRepo.Update(ctx, image); err != nil {
		uc.logger.Error("Failed to update image", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to update image: %w", err)
	}

	uc.logger.Info("Image updated successfully", slog.String("image_id", imageID.String()))
	return image, nil
}

// DeleteImage deletes an image
func (uc *ProductImageUseCase) DeleteImage(ctx context.Context, imageID uuid.UUID) error {
	uc.logger.Info("Deleting image", slog.String("image_id", imageID.String()))

	// Get image to check if it's primary
	image, err := uc.GetImageByID(ctx, imageID)
	if err != nil {
		return err
	}

	// If it's a primary image, reassign primary to another image
	if image.IsPrimary {
		if err := uc.reassignPrimaryImage(ctx, image); err != nil {
			uc.logger.Error("Failed to reassign primary image", slog.String("error", err.Error()))
			return fmt.Errorf("failed to reassign primary image: %w", err)
		}
	}

	// Delete image
	if err := uc.imageRepo.Delete(ctx, imageID); err != nil {
		uc.logger.Error("Failed to delete image", slog.String("error", err.Error()))
		return fmt.Errorf("failed to delete image: %w", err)
	}

	uc.logger.Info("Image deleted successfully", slog.String("image_id", imageID.String()))
	return nil
}

// GetProductImages retrieves all images for a product
func (uc *ProductImageUseCase) GetProductImages(ctx context.Context, productID uuid.UUID) ([]*entity.ProductImage, error) {
	uc.logger.Info("Getting product images", slog.String("product_id", productID.String()))

	include := &repository.ProductImageInclude{
		Product: true,
	}

	images, err := uc.imageRepo.GetByProduct(ctx, productID, include)
	if err != nil {
		uc.logger.Error("Failed to get product images",
			slog.String("product_id", productID.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get product images: %w", err)
	}

	return images, nil
}

// GetVariantImages retrieves all images for a variant
func (uc *ProductImageUseCase) GetVariantImages(ctx context.Context, variantID uuid.UUID) ([]*entity.ProductImage, error) {
	uc.logger.Info("Getting variant images", slog.String("variant_id", variantID.String()))

	include := &repository.ProductImageInclude{
		Variant: true,
	}

	images, err := uc.imageRepo.GetByVariant(ctx, variantID, include)
	if err != nil {
		uc.logger.Error("Failed to get variant images",
			slog.String("variant_id", variantID.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get variant images: %w", err)
	}

	return images, nil
}

// SetPrimaryImage sets an image as primary for its product or variant
func (uc *ProductImageUseCase) SetPrimaryImage(ctx context.Context, req SetPrimaryImageRequest) error {
	uc.logger.Info("Setting primary image",
		slog.String("product_id", req.ProductID.String()),
		slog.Any("variant_id", req.VariantID),
		slog.String("image_id", req.ImageID.String()))

	// Get the image to verify it exists and belongs to the product/variant
	image, err := uc.GetImageByID(ctx, req.ImageID)
	if err != nil {
		return err
	}

	// Verify image belongs to the specified product
	if image.ProductID != req.ProductID {
		return fmt.Errorf("image does not belong to the specified product")
	}

	// Verify variant relationship if specified
	if req.VariantID != nil {
		if image.VariantID == nil || *image.VariantID != *req.VariantID {
			return fmt.Errorf("image does not belong to the specified variant")
		}
	} else if image.VariantID != nil {
		return fmt.Errorf("cannot set variant image as primary for product")
	}

	// Unset current primary images
	if err := uc.imageRepo.UnsetPrimaryImages(ctx, req.ProductID, req.VariantID); err != nil {
		uc.logger.Error("Failed to unset current primary images", slog.String("error", err.Error()))
		return fmt.Errorf("failed to unset current primary images: %w", err)
	}

	// Set new primary image
	if req.VariantID != nil {
		err = uc.imageRepo.SetVariantPrimaryImage(ctx, *req.VariantID, req.ImageID)
	} else {
		err = uc.imageRepo.SetProductPrimaryImage(ctx, req.ProductID, req.ImageID)
	}

	if err != nil {
		uc.logger.Error("Failed to set primary image", slog.String("error", err.Error()))
		return fmt.Errorf("failed to set primary image: %w", err)
	}

	uc.logger.Info("Primary image set successfully")
	return nil
}

// ReorderImages updates the sort order of images
func (uc *ProductImageUseCase) ReorderImages(ctx context.Context, req ReorderImagesRequest) error {
	uc.logger.Info("Reordering images",
		slog.String("product_id", req.ProductID.String()),
		slog.Any("variant_id", req.VariantID),
		slog.Int("image_count", len(req.ImageOrders)))

	// Validate that all images belong to the product/variant
	if err := uc.validateImageOwnership(ctx, req.ProductID, req.VariantID, req.ImageOrders); err != nil {
		uc.logger.Error("Image ownership validation failed", slog.String("error", err.Error()))
		return fmt.Errorf("image ownership validation failed: %w", err)
	}

	// Reorder images
	if err := uc.imageRepo.ReorderImages(ctx, req.ProductID, req.VariantID, req.ImageOrders); err != nil {
		uc.logger.Error("Failed to reorder images", slog.String("error", err.Error()))
		return fmt.Errorf("failed to reorder images: %w", err)
	}

	uc.logger.Info("Images reordered successfully")
	return nil
}

// BulkUploadImages uploads multiple images for a product/variant
func (uc *ProductImageUseCase) BulkUploadImages(ctx context.Context, req BulkImageUploadRequest) ([]*entity.ProductImage, error) {
	uc.logger.Info("Bulk uploading images",
		slog.String("product_id", req.ProductID.String()),
		slog.Any("variant_id", req.VariantID),
		slog.Int("image_count", len(req.Images)))

	// Validate bulk upload constraints
	if err := uc.validateBulkUpload(ctx, req); err != nil {
		uc.logger.Error("Bulk upload validation failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("bulk upload validation failed: %w", err)
	}

	var images []*entity.ProductImage
	var errors []error

	// Process each image
	for i, imgReq := range req.Images {
		imgReq.ProductID = req.ProductID
		imgReq.VariantID = req.VariantID

		image, err := uc.CreateImage(ctx, imgReq)
		if err != nil {
			uc.logger.Error("Failed to create image in bulk upload",
				slog.Int("image_index", i),
				slog.String("error", err.Error()))
			errors = append(errors, fmt.Errorf("image %d: %w", i+1, err))
			continue
		}

		images = append(images, image)
	}

	// If there were any errors, return them
	if len(errors) > 0 {
		return images, fmt.Errorf("bulk upload completed with errors: %v", errors)
	}

	uc.logger.Info("Bulk image upload completed successfully",
		slog.Int("uploaded_count", len(images)))
	return images, nil
}

// GetPrimaryImage gets the primary image for a product or variant
func (uc *ProductImageUseCase) GetPrimaryImage(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID) (*entity.ProductImage, error) {
	uc.logger.Info("Getting primary image",
		slog.String("product_id", productID.String()),
		slog.Any("variant_id", variantID))

	include := &repository.ProductImageInclude{
		Product: true,
		Variant: variantID != nil,
	}

	var image *entity.ProductImage
	var err error

	if variantID != nil {
		image, err = uc.imageRepo.GetVariantPrimaryImage(ctx, *variantID, include)
	} else {
		image, err = uc.imageRepo.GetProductPrimaryImage(ctx, productID, include)
	}

	if err != nil {
		uc.logger.Error("Failed to get primary image",
			slog.String("product_id", productID.String()),
			slog.Any("variant_id", variantID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get primary image: %w", err)
	}

	return image, nil
}

// UpdateImageCloudinaryInfo updates Cloudinary information for an image
func (uc *ProductImageUseCase) UpdateImageCloudinaryInfo(ctx context.Context, imageID uuid.UUID, cloudinaryID, cloudinaryURL *string) error {
	uc.logger.Info("Updating image Cloudinary info", slog.String("image_id", imageID.String()))

	if err := uc.imageRepo.UpdateCloudinaryInfo(ctx, imageID, cloudinaryID, cloudinaryURL, nil, nil, nil); err != nil {
		uc.logger.Error("Failed to update Cloudinary info", slog.String("error", err.Error()))
		return fmt.Errorf("failed to update Cloudinary info: %w", err)
	}

	uc.logger.Info("Cloudinary info updated successfully")
	return nil
}

// ValidateImageURL validates if an image URL is accessible and valid
func (uc *ProductImageUseCase) ValidateImageURL(ctx context.Context, imageURL string) error {
	uc.logger.Info("Validating image URL", slog.String("url", imageURL))

	// Basic URL validation
	if _, err := url.Parse(imageURL); err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check accessibility (optional - can be disabled in production for performance)
	if err := uc.imageRepo.CheckImageAccessibility(ctx, imageURL); err != nil {
		uc.logger.Warn("Image accessibility check failed",
			slog.String("url", imageURL),
			slog.String("error", err.Error()))
		// Don't fail validation on accessibility check - just log warning
	}

	return nil
}

// Helper methods

func (uc *ProductImageUseCase) validateCreateImageRequest(ctx context.Context, req CreateImageRequest) error {
	// Validate product exists
	exists, err := uc.productRepo.Exists(ctx, req.ProductID)
	if err != nil {
		return fmt.Errorf("failed to check product existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("product with ID %s does not exist", req.ProductID)
	}

	// Validate variant exists if specified
	if req.VariantID != nil {
		exists, err := uc.variantRepo.Exists(ctx, *req.VariantID)
		if err != nil {
			return fmt.Errorf("failed to check variant existence: %w", err)
		}
		if !exists {
			return fmt.Errorf("variant with ID %s does not exist", *req.VariantID)
		}
	}

	// Validate URL
	if err := uc.ValidateImageURL(ctx, req.ImageURL); err != nil {
		return fmt.Errorf("invalid image URL: %w", err)
	}

	// Validate MIME type if provided
	if req.MimeType != nil {
		if !uc.isValidMimeType(*req.MimeType) {
			return fmt.Errorf("unsupported MIME type: %s", *req.MimeType)
		}
	}

	// Validate dimensions if provided
	if req.Width != nil && req.Height != nil {
		if err := uc.imageRepo.ValidateImageDimensions(ctx, *req.Width, *req.Height); err != nil {
			return fmt.Errorf("invalid image dimensions: %w", err)
		}
	}

	return nil
}

func (uc *ProductImageUseCase) applyImageFields(image *entity.ProductImage, req CreateImageRequest) {
	if req.CloudinaryURL != nil {
		image.CloudinaryURL = req.CloudinaryURL
	}
	if req.CloudinaryPublicID != nil {
		image.CloudinaryID = req.CloudinaryPublicID
	}
	if req.AltText != nil {
		image.AltText = req.AltText
	}
	if req.Width != nil {
		image.Width = req.Width
	}
	if req.Height != nil {
		image.Height = req.Height
	}
	if req.FileSize != nil {
		image.FileSize = req.FileSize
	}
	if req.MimeType != nil {
		image.MimeType = req.MimeType
	}
	if req.IsPrimary != nil {
		image.IsPrimary = *req.IsPrimary
	}
	if req.SortOrder != nil {
		image.SortOrder = *req.SortOrder
	}
}

func (uc *ProductImageUseCase) applyImageUpdates(image *entity.ProductImage, req UpdateImageRequest) {
	if req.ImageURL != nil {
		image.ImageURL = *req.ImageURL
	}
	if req.CloudinaryURL != nil {
		image.CloudinaryURL = req.CloudinaryURL
	}
	if req.CloudinaryPublicID != nil {
		image.CloudinaryID = req.CloudinaryPublicID
	}
	if req.AltText != nil {
		image.AltText = req.AltText
	}
	if req.Width != nil {
		image.Width = req.Width
	}
	if req.Height != nil {
		image.Height = req.Height
	}
	if req.FileSize != nil {
		image.FileSize = req.FileSize
	}
	if req.MimeType != nil {
		image.MimeType = req.MimeType
	}
	if req.IsPrimary != nil {
		image.IsPrimary = *req.IsPrimary
	}
	if req.SortOrder != nil {
		image.SortOrder = *req.SortOrder
	}
}

func (uc *ProductImageUseCase) validateImageConstraints(ctx context.Context, image *entity.ProductImage) error {
	options := DefaultImageValidationOptions()

	// Check file size
	if image.FileSize != nil && *image.FileSize > options.MaxFileSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size of %d bytes", *image.FileSize, options.MaxFileSize)
	}

	// Check image count
	currentCount, err := uc.imageRepo.GetImageCount(ctx, image.ProductID, image.VariantID)
	if err != nil {
		return fmt.Errorf("failed to get current image count: %w", err)
	}
	if currentCount >= options.MaxImageCount {
		return fmt.Errorf("maximum image count of %d reached", options.MaxImageCount)
	}

	// Check dimensions
	if image.Width != nil && image.Height != nil {
		if *image.Width < options.MinWidth || *image.Width > options.MaxWidth {
			return fmt.Errorf("image width %d is outside allowed range %d-%d", *image.Width, options.MinWidth, options.MaxWidth)
		}
		if *image.Height < options.MinHeight || *image.Height > options.MaxHeight {
			return fmt.Errorf("image height %d is outside allowed range %d-%d", *image.Height, options.MinHeight, options.MaxHeight)
		}
	}

	// Check alt text requirement
	if options.RequireAltText && (image.AltText == nil || strings.TrimSpace(*image.AltText) == "") {
		return fmt.Errorf("alt text is required")
	}

	return nil
}

func (uc *ProductImageUseCase) handlePrimaryImageAssignment(ctx context.Context, image *entity.ProductImage) error {
	// Unset current primary images for the same product/variant
	return uc.imageRepo.UnsetPrimaryImages(ctx, image.ProductID, image.VariantID)
}

func (uc *ProductImageUseCase) reassignPrimaryImage(ctx context.Context, deletedImage *entity.ProductImage) error {
	// Get other images for the same product/variant
	var images []*entity.ProductImage
	var err error

	include := &repository.ProductImageInclude{}

	if deletedImage.VariantID != nil {
		images, err = uc.imageRepo.GetByVariant(ctx, *deletedImage.VariantID, include)
	} else {
		images, err = uc.imageRepo.GetByProduct(ctx, deletedImage.ProductID, include)
	}

	if err != nil {
		return fmt.Errorf("failed to get other images: %w", err)
	}

	// Find the next image to set as primary (lowest sort order)
	var nextPrimary *entity.ProductImage
	for _, img := range images {
		if img.ID != deletedImage.ID {
			if nextPrimary == nil || img.SortOrder < nextPrimary.SortOrder {
				nextPrimary = img
			}
		}
	}

	// Set the next image as primary if found
	if nextPrimary != nil {
		if deletedImage.VariantID != nil {
			return uc.imageRepo.SetVariantPrimaryImage(ctx, *deletedImage.VariantID, nextPrimary.ID)
		} else {
			return uc.imageRepo.SetProductPrimaryImage(ctx, deletedImage.ProductID, nextPrimary.ID)
		}
	}

	return nil
}

func (uc *ProductImageUseCase) validateImageOwnership(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID, imageOrders []repository.ImageOrder) error {
	// Get all image IDs to validate
	imageIDs := make([]uuid.UUID, len(imageOrders))
	for i, order := range imageOrders {
		imageIDs[i] = order.ImageID
	}

	// Get images by IDs
	include := &repository.ProductImageInclude{}
	images, err := uc.imageRepo.GetByIDs(ctx, imageIDs, include)
	if err != nil {
		return fmt.Errorf("failed to get images by IDs: %w", err)
	}

	// Verify all images belong to the specified product/variant
	for _, image := range images {
		if image.ProductID != productID {
			return fmt.Errorf("image %s does not belong to product %s", image.ID, productID)
		}

		if variantID != nil {
			if image.VariantID == nil || *image.VariantID != *variantID {
				return fmt.Errorf("image %s does not belong to variant %s", image.ID, *variantID)
			}
		} else if image.VariantID != nil {
			return fmt.Errorf("image %s belongs to a variant, not the product", image.ID)
		}
	}

	return nil
}

func (uc *ProductImageUseCase) validateBulkUpload(ctx context.Context, req BulkImageUploadRequest) error {
	if len(req.Images) == 0 {
		return fmt.Errorf("no images provided")
	}

	options := DefaultImageValidationOptions()
	if len(req.Images) > options.MaxImageCount {
		return fmt.Errorf("too many images: %d, maximum allowed: %d", len(req.Images), options.MaxImageCount)
	}

	// Check current image count
	currentCount, err := uc.imageRepo.GetImageCount(ctx, req.ProductID, req.VariantID)
	if err != nil {
		return fmt.Errorf("failed to get current image count: %w", err)
	}

	if currentCount+len(req.Images) > options.MaxImageCount {
		return fmt.Errorf("total image count would exceed maximum of %d", options.MaxImageCount)
	}

	// Validate each image request
	for i, imgReq := range req.Images {
		if err := uc.ValidateImageURL(ctx, imgReq.ImageURL); err != nil {
			return fmt.Errorf("invalid URL for image %d: %w", i+1, err)
		}

		if imgReq.MimeType != nil && !uc.isValidMimeType(*imgReq.MimeType) {
			return fmt.Errorf("unsupported MIME type for image %d: %s", i+1, *imgReq.MimeType)
		}
	}

	return nil
}

func (uc *ProductImageUseCase) isValidMimeType(mimeType string) bool {
	// Normalize MIME type
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))

	// Extract media type without parameters
	mediaType, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		return false
	}

	// Check if it's a valid image MIME type
	allowedTypes := DefaultImageValidationOptions().AllowedMimeTypes
	for _, allowed := range allowedTypes {
		if mediaType == allowed {
			return true
		}
	}

	return false
}
