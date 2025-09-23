package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// RepositoryFactory provides a centralized way to create repository instances
type RepositoryFactory struct {
	db *sqlx.DB
}

// RepositoryConfig holds configuration for repository creation
type RepositoryConfig struct {
	DatabaseType string // "postgresql", "mysql", etc.
	// Add other configuration options as needed
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(db *sqlx.DB) *RepositoryFactory {
	return &RepositoryFactory{
		db: db,
	}
}

// ProductRepository creates a new product repository instance
func (f *RepositoryFactory) ProductRepository() repository.ProductRepository {
	return NewPostgreSQLProductRepository(f.db)
}

// ProductCategoryRepository creates a new product category repository instance
func (f *RepositoryFactory) ProductCategoryRepository() repository.ProductCategoryRepository {
	return NewPostgreSQLProductCategoryRepository(f.db)
}

// ProductVariantOptionRepository creates a new product variant option repository instance
func (f *RepositoryFactory) ProductVariantOptionRepository() repository.ProductVariantOptionRepository {
	return NewPostgreSQLProductVariantOptionRepository(f.db)
}

// ProductVariantRepository creates a new product variant repository instance
func (f *RepositoryFactory) ProductVariantRepository() repository.ProductVariantRepository {
	return NewPostgreSQLProductVariantRepository(f.db)
}

// ProductImageRepository creates a new product image repository instance
func (f *RepositoryFactory) ProductImageRepository() repository.ProductImageRepository {
	return NewPostgreSQLProductImageRepository(f.db)
}

// RepositoryContainer holds all repository instances for dependency injection
type RepositoryContainer struct {
	ProductRepository              repository.ProductRepository
	ProductCategoryRepository      repository.ProductCategoryRepository
	ProductVariantOptionRepository repository.ProductVariantOptionRepository
	ProductVariantRepository       repository.ProductVariantRepository
	ProductImageRepository         repository.ProductImageRepository
}

// NewRepositoryContainer creates a new repository container with all repositories initialized
func NewRepositoryContainer(factory *RepositoryFactory) *RepositoryContainer {
	return &RepositoryContainer{
		ProductRepository:              factory.ProductRepository(),
		ProductCategoryRepository:      factory.ProductCategoryRepository(),
		ProductVariantOptionRepository: factory.ProductVariantOptionRepository(),
		ProductVariantRepository:       factory.ProductVariantRepository(),
		ProductImageRepository:         factory.ProductImageRepository(),
	}
}

// RepositoryProvider is an interface for providing repository instances
// This can be useful for dependency injection frameworks
type RepositoryProvider interface {
	GetProductRepository() repository.ProductRepository
	GetProductCategoryRepository() repository.ProductCategoryRepository
	GetProductVariantOptionRepository() repository.ProductVariantOptionRepository
	GetProductVariantRepository() repository.ProductVariantRepository
	GetProductImageRepository() repository.ProductImageRepository
}

// Ensure RepositoryContainer implements RepositoryProvider
var _ RepositoryProvider = (*RepositoryContainer)(nil)

// RepositoryProvider implementation
func (c *RepositoryContainer) GetProductRepository() repository.ProductRepository {
	return c.ProductRepository
}

func (c *RepositoryContainer) GetProductCategoryRepository() repository.ProductCategoryRepository {
	return c.ProductCategoryRepository
}

func (c *RepositoryContainer) GetProductVariantOptionRepository() repository.ProductVariantOptionRepository {
	return c.ProductVariantOptionRepository
}

func (c *RepositoryContainer) GetProductVariantRepository() repository.ProductVariantRepository {
	return c.ProductVariantRepository
}

func (c *RepositoryContainer) GetProductImageRepository() repository.ProductImageRepository {
	return c.ProductImageRepository
}

// RepositoryType represents the type of repository implementation
type RepositoryType string

const (
	RepositoryTypePostgreSQL RepositoryType = "postgresql"
	RepositoryTypeMySQL      RepositoryType = "mysql"
	RepositoryTypeMongoDB    RepositoryType = "mongodb"
	RepositoryTypeInMemory   RepositoryType = "memory"
)

// AdvancedRepositoryFactory supports different repository implementations
type AdvancedRepositoryFactory struct {
	db       *sqlx.DB
	repoType RepositoryType
	config   *RepositoryConfig
}

// NewAdvancedRepositoryFactory creates a new advanced repository factory
func NewAdvancedRepositoryFactory(db *sqlx.DB, repoType RepositoryType, config *RepositoryConfig) *AdvancedRepositoryFactory {
	return &AdvancedRepositoryFactory{
		db:       db,
		repoType: repoType,
		config:   config,
	}
}

// CreateProductRepository creates a product repository based on the configured type
func (f *AdvancedRepositoryFactory) CreateProductRepository() (repository.ProductRepository, error) {
	switch f.repoType {
	case RepositoryTypePostgreSQL:
		return NewPostgreSQLProductRepository(f.db), nil
	case RepositoryTypeMySQL:
		// TODO: Implement MySQL repository when needed
		return nil, fmt.Errorf("MySQL repository not implemented yet")
	case RepositoryTypeMongoDB:
		// TODO: Implement MongoDB repository when needed
		return nil, fmt.Errorf("MongoDB repository not implemented yet")
	case RepositoryTypeInMemory:
		// TODO: Implement in-memory repository for testing
		return nil, fmt.Errorf("in-memory repository not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported repository type: %s", f.repoType)
	}
}

// CreateProductCategoryRepository creates a product category repository based on the configured type
func (f *AdvancedRepositoryFactory) CreateProductCategoryRepository() (repository.ProductCategoryRepository, error) {
	switch f.repoType {
	case RepositoryTypePostgreSQL:
		return NewPostgreSQLProductCategoryRepository(f.db), nil
	default:
		return nil, fmt.Errorf("unsupported repository type: %s", f.repoType)
	}
}

// CreateProductVariantOptionRepository creates a product variant option repository based on the configured type
func (f *AdvancedRepositoryFactory) CreateProductVariantOptionRepository() (repository.ProductVariantOptionRepository, error) {
	switch f.repoType {
	case RepositoryTypePostgreSQL:
		return NewPostgreSQLProductVariantOptionRepository(f.db), nil
	default:
		return nil, fmt.Errorf("unsupported repository type: %s", f.repoType)
	}
}

// CreateProductVariantRepository creates a product variant repository based on the configured type
func (f *AdvancedRepositoryFactory) CreateProductVariantRepository() (repository.ProductVariantRepository, error) {
	switch f.repoType {
	case RepositoryTypePostgreSQL:
		return NewPostgreSQLProductVariantRepository(f.db), nil
	default:
		return nil, fmt.Errorf("unsupported repository type: %s", f.repoType)
	}
}

// CreateProductImageRepository creates a product image repository based on the configured type
func (f *AdvancedRepositoryFactory) CreateProductImageRepository() (repository.ProductImageRepository, error) {
	switch f.repoType {
	case RepositoryTypePostgreSQL:
		return NewPostgreSQLProductImageRepository(f.db), nil
	default:
		return nil, fmt.Errorf("unsupported repository type: %s", f.repoType)
	}
}

// CreateAllRepositories creates all repositories and returns them in a container
func (f *AdvancedRepositoryFactory) CreateAllRepositories() (*RepositoryContainer, error) {
	productRepo, err := f.CreateProductRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to create product repository: %w", err)
	}

	categoryRepo, err := f.CreateProductCategoryRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to create product category repository: %w", err)
	}

	variantOptionRepo, err := f.CreateProductVariantOptionRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to create product variant option repository: %w", err)
	}

	variantRepo, err := f.CreateProductVariantRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to create product variant repository: %w", err)
	}

	imageRepo, err := f.CreateProductImageRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to create product image repository: %w", err)
	}

	return &RepositoryContainer{
		ProductRepository:              productRepo,
		ProductCategoryRepository:      categoryRepo,
		ProductVariantOptionRepository: variantOptionRepo,
		ProductVariantRepository:       variantRepo,
		ProductImageRepository:         imageRepo,
	}, nil
}

// RepositoryInterface verification - compile-time check
func VerifyRepositoryInterfaces() {
	// This function ensures all our repositories implement their interfaces correctly
	// It will cause compilation errors if interfaces are not properly implemented

	var db *sqlx.DB // This is just for interface verification

	var _ repository.ProductRepository = NewPostgreSQLProductRepository(db)
	var _ repository.ProductCategoryRepository = NewPostgreSQLProductCategoryRepository(db)
	var _ repository.ProductVariantOptionRepository = NewPostgreSQLProductVariantOptionRepository(db)
	var _ repository.ProductVariantRepository = NewPostgreSQLProductVariantRepository(db)
	var _ repository.ProductImageRepository = NewPostgreSQLProductImageRepository(db)
}
