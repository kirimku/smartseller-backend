Service and Usecase Separation in Kirimku Backend
Overview
Kirimku Backend follows a clean architecture approach with a clear separation of concerns between usecases and services. This document explains the design philosophy, responsibilities, and implementation details of each layer.

Architectural Layers
Domain Layer
Contains business entities and repository interfaces
Defines the core business rules independent of any external technology or framework
Application Layer
Contains usecases and services
Implements business logic and coordinates between domain entities and external services
Infrastructure Layer
Contains repository implementations and external service adapters
Connects the application to databases, APIs, and other external resources
Interface Layer
Contains API handlers, controllers, and DTOs
Handles HTTP requests and translates between external formats and internal models
Usecases vs Services
Usecases
Usecases represent the core business logic of the application.

Responsibilities:
Implement domain-specific business rules
Coordinate operations between multiple repositories
Process business transactions
Enforce business constraints and validations
Maintain the integrity of domain entities
Characteristics:
Independent of external services: Should work with mock repositories for testing
Domain-centric: Work directly with domain entities
Repository-focused: Interact with repositories to access and manipulate data
Business process oriented: Represent complete business processes
Examples:
TransactionUsecase: Handles the complete lifecycle of a shipping transaction
CashbackUsecase: Manages cashback calculation, processing, and distribution
UserUsecase: Handles user management, authentication, and authorization
Services
Services act as integration and coordination layers that connect usecases with external systems.

Responsibilities:
Integrate with external APIs and third-party services
Transform data between external and internal formats
Provide reusable functionality to multiple usecases
Handle cross-cutting concerns like caching, logging, etc.
Coordinate between multiple usecases for complex operations
Characteristics:
Integration-focused: Connect internal systems with external resources
Stateless: Generally don't maintain state
Transformational: Often transform between data formats
Technical in nature: Address technical concerns rather than business logic
Examples:
PricingService: Calculates shipping costs using courier pricing rules
NotificationService: Sends notifications through various channels
LocationService: Provides location validation and geocoding
CourierService: Interfaces with external courier APIs
Implementation Details
Usecase Implementation
Usecases are implemented as Go interfaces with concrete implementations:
```
type TransactionUsecase interface {
    CreateTransaction(ctx context.Context, transaction *entity.Transaction) (*entity.Transaction, float64, error)
    // Other transaction-related methods...
}

type TransactionUsecaseImpl struct {
    transactionRepo repository.TransactionRepository
    userRepo        repository.UserRepository
    cashbackUsecase CashbackUsecase
    pricingService  PricingService
    // Dependencies on repositories and services
}
```

Service Implementation
Services are also implemented as Go interfaces with concrete implementations:
```
type PricingService interface {
    CalculateShippingCost(ctx context.Context, transaction *entity.Transaction) (float64, error)
    // Other pricing-related methods...
}

type PricingServiceImpl struct { . .
    courierRatesRepo repository.CourierRatesRepository
    redisClient      cache.RedisClient
    // Dependencies on repositories and other services
}
```

Dependency Flow
The dependency flow is unidirectional:

Usecases can depend on repositories and services
Services can depend on repositories and other services
Neither should depend on controllers/handlers
Controllers/handlers depend on usecases and services
This ensures that the core business logic (usecases) remains clean and independent of implementation details.

Best Practices
Keep usecases focused on business logic:

No HTTP, SQL, or external API calls directly in usecases
Use repositories and services for external interactions
Make services reusable:

Design services to be used by multiple usecases
Keep services focused on a single responsibility
Inject dependencies:

Use dependency injection for repositories and services
Avoid creating dependencies inside usecases or services
Testing:

Usecases should be testable with mock repositories and services
Services should handle error cases from external systems gracefully
Example Flow: Creating a Transaction
API Handler receives HTTP request and transforms to domain entity
Transaction Usecase validates the transaction and applies business rules
Pricing Service calculates shipping costs from courier rates
Transaction Usecase creates cost components for insurance, cashback, etc.
Transaction Repository persists the transaction to database
Cashback Usecase processes cashback for the transaction
API Handler transforms domain entity back to HTTP response
This separation ensures that business logic remains clean and maintainable while external integrations can evolve independently.