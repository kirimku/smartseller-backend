# Interfaces Layer

## Overview

The `interfaces` folder implements the boundary between the core domain of the application and the outside world. In Clean Architecture (or Hexagonal Architecture), this layer serves as the adapter layer that translates between external systems and the application's core business logic.

## Why "Interfaces"?

The name "interfaces" reflects the role this layer plays in the architecture:

1. **Boundary Definition**: It defines how external actors interact with our application.
2. **Implementation of Domain Interfaces**: It provides concrete implementations of interfaces defined in the domain layer.
3. **Adapter Pattern**: It adapts external concerns to internal abstractions and vice versa.
4. **Input/Output Ports**: It acts as both "input ports" (API handlers) and "output ports" (repositories).

## Structure

This layer contains several subdirectories:

- **`api/`**: HTTP API-related code
  - **`handler/`**: Request handlers that convert HTTP requests to domain operations
  - **`middleware/`**: HTTP middleware components (auth, logging, error handling)
  - **`router/`**: Route definitions and setup
  - **`validator/`**: Request validation logic
- **`mapper/`**: Maps between API DTOs and domain entities

## Benefits of This Architecture

By keeping the interfaces layer separate from both domain logic and infrastructure concerns:

1. **Testability**: Business logic can be tested without HTTP or database dependencies
2. **Flexibility**: Easy to change delivery mechanisms (e.g., from REST to gRPC)
3. **Maintainability**: Clear boundaries between different concerns
4. **Dependency Direction**: Dependencies flow inward, following Dependency Inversion Principle

## Design Principles

1. **Single Responsibility**: Each component handles a single concern
2. **Dependency Inversion**: High-level modules don't depend on low-level modules
3. **Interface Segregation**: Clients depend only on interfaces they actually use
4. **Decoupling**: External concerns are separated from business logic

## Best Practices

When working in this layer:

1. Keep handlers thin - delegate to use cases for business logic
2. Ensure proper error handling and input validation
3. Map between DTOs and domain entities
4. Don't leak domain entities to the outside world
5. Don't include business logic in handlers

## Relation to Other Layers

- **Domain Layer**: Contains business entities and repository interfaces
- **Application Layer**: Contains use cases and business rules
- **Infrastructure Layer**: Contains concrete implementations of repositories and external services

This architecture follows the principles laid out by Robert C. Martin in "Clean Architecture" and Alistair Cockburn's "Hexagonal Architecture" (Ports and Adapters).