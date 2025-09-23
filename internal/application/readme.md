Domain Layer
The domain folder contains the core business logic and rules of your application:

Entities: These are the core business objects (like entity.Address)
Repository Interfaces: These define how to access data (like repository.AddressRepository)
Domain Services: These define core business operations (like service.AddressService)
The domain layer is independent of external concerns and should contain pure business logic without dependencies on frameworks, databases, or UI.

Key Differences
Purpose:

Domain: Core business rules and entities
Application: Use cases and coordination
Dependencies:

Domain: No dependencies on external frameworks
Application: Can depend on frameworks and adapters
Knowledge:

Domain: Knows nothing about external concerns (HTTP, DB)
Application: Knows about both domain and external interfaces
In your code, you can see this pattern clearly:

domain/service/AddressService defines the interface with business operations
application/service/AddressService implements that interface, handling DTO conversion and coordinating with repositories
This separation helps maintain clean architecture and makes your code more testable and maintainable by isolating core business logic from infrastructure concerns.