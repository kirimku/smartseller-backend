# SmartSeller Backend

A modern, cloud-native e-commerce management platform backend built with Go, designed to empower sellers with comprehensive business management tools.

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)]()
[![Go Version](https://img.shields.io/badge/go-1.24.0-blue)]()
[![Architecture](https://img.shields.io/badge/architecture-microservices-orange)]()
[![License](https://img.shields.io/badge/license-MIT-green)]()

## 🚀 Overview

SmartSeller is a comprehensive e-commerce business management platform that provides sellers with everything they need to build, manage, and scale their online business. From individual entrepreneurs to enterprise retailers, SmartSeller offers an integrated solution for modern e-commerce operations.

## ✨ Key Features

### 🔐 Authentication & User Management
- Secure user registration and login
- JWT-based authentication with refresh tokens
- OAuth integration (Google, Facebook, etc.)
- Role-based access control (RBAC)
- Password reset and account recovery

### 🛍️ E-commerce Ready Foundation
- Multi-tenant architecture
- User tiers: Basic, Premium, Pro, Enterprise
- User types: Individual, Business, Enterprise
- Scalable API architecture for future features

### 🏗️ Technical Excellence
- **Clean Architecture**: Domain-driven design with clear separation of concerns
- **Microservices Ready**: Modular architecture for independent scaling
- **Cloud Native**: Containerized and Kubernetes-ready
- **API First**: RESTful APIs with comprehensive documentation
- **Security First**: Enterprise-grade security with encryption and audit trails

## 🎯 Planned Features

### Phase 1: Product Management
- [ ] Product catalog with unlimited variants
- [ ] Category management and organization
- [ ] Inventory tracking and alerts
- [ ] Digital asset management with CDN
- [ ] Bulk import/export capabilities

### Phase 2: Order Management
- [ ] Shopping cart and checkout flow
- [ ] Order processing automation
- [ ] Payment gateway integration
- [ ] Invoice generation and management
- [ ] Return and refund processing

### Phase 3: Customer Management
- [ ] 360-degree customer profiles
- [ ] Communication history tracking
- [ ] Customer segmentation tools
- [ ] Support ticket system
- [ ] Customer analytics and insights

### Phase 4: Channel Management
- [ ] Marketplace integration (Shopee, Tokopedia, etc.)
- [ ] Social commerce (Instagram, Facebook)
- [ ] Custom storefront deployment
- [ ] Omnichannel inventory sync
- [ ] Multi-channel order management

### Phase 5: Marketing & Loyalty
- [ ] Voucher and discount management
- [ ] Loyalty points and rewards system
- [ ] Email marketing automation
- [ ] Customer segmentation campaigns
- [ ] A/B testing framework

### Phase 6: Analytics & Intelligence
- [ ] Real-time sales dashboards
- [ ] Business intelligence reports
- [ ] Predictive analytics
- [ ] Inventory optimization
- [ ] Financial reporting

## 🏛️ Architecture

SmartSeller follows clean architecture principles with clear domain boundaries:

```
┌─────────────────────────────────────────────────────────────┐
│                    API Layer (Gin HTTP)                     │
├─────────────────────────────────────────────────────────────┤
│                 Application Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Use Cases │  │   Services  │  │    DTOs     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│                   Domain Layer                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  Entities   │  │ Repositories│  │   Services  │         │
│  │             │  │ (Interfaces)│  │             │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│                Infrastructure Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  Database   │  │   External  │  │    Cache    │         │
│  │ (PostgreSQL)│  │   Services  │  │   (Redis)   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

## 🛠️ Technology Stack

### Backend
- **Language**: Go 1.24.0+
- **Framework**: Gin HTTP Framework
- **Database**: PostgreSQL with migrations support
- **Cache**: Redis for sessions and caching
- **Authentication**: JWT with refresh tokens
- **Documentation**: OpenAPI/Swagger

### Infrastructure
- **Containerization**: Docker & Kubernetes
- **Cloud**: Multi-cloud support (AWS, GCP, Azure)
- **Monitoring**: Prometheus, Grafana, Jaeger
- **CI/CD**: GitHub Actions
- **Security**: TLS encryption, RBAC, audit logs

## 🚦 Getting Started

### Prerequisites
- Go 1.24.0 or higher
- PostgreSQL 13+
- Redis 6+
- Docker (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/kirimku/smartseller-backend.git
   cd smartseller-backend
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run database migrations**
   ```bash
   make migrate-up
   ```

5. **Start the server**
   ```bash
   make run
   ```

The API will be available at `http://localhost:8080`

### Docker Setup

```bash
# Build and run with Docker Compose
docker-compose up --build

# Or build individual image
docker build -t smartseller-backend .
docker run -p 8080:8080 smartseller-backend
```

## 📚 API Documentation

Once the server is running, you can access:

- **API Documentation**: `http://localhost:8080/docs`
- **Health Check**: `http://localhost:8080/health`
- **API Endpoints**: `http://localhost:8080/api/v1/*`

### Authentication Endpoints

```bash
# Register a new user
POST /api/v1/auth/register
{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "password": "securepassword",
  "user_type": "individual",
  "accept_terms": true
}

# Login
POST /api/v1/auth/login
{
  "email_or_phone": "john@example.com",
  "password": "securepassword"
}

# Refresh token
POST /api/v1/auth/refresh
{
  "refresh_token": "your_refresh_token"
}
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration

# Run specific test
go test ./internal/application/usecase -v
```

## 🚀 Deployment

### Production Deployment

1. **Build the application**
   ```bash
   make build
   ```

2. **Deploy with Docker**
   ```bash
   docker build -t smartseller-backend:latest .
   docker push your-registry/smartseller-backend:latest
   ```

3. **Deploy to Kubernetes**
   ```bash
   kubectl apply -f k8s/
   ```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_NAME` | Database name | `smartseller` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | - |
| `REDIS_URL` | Redis connection URL | `redis://localhost:6379` |
| `JWT_SECRET` | JWT signing secret | - |
| `SESSION_KEY` | Session encryption key | - |

## 🤝 Contributing

We welcome contributions! Please read our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [Business Plan](SMARTSELLER_BUSINESS_PLAN.md) - Comprehensive business strategy
- [Technical Architecture](TECHNICAL_ARCHITECTURE.md) - Detailed technical design
- [Cleanup Summary](CLEANUP_SUMMARY.md) - Migration from kirimku to SmartSeller
- [API Documentation](http://localhost:8080/docs) - OpenAPI/Swagger docs

## 📞 Support

- **Documentation**: Check our [docs](./docs/)
- **Issues**: [GitHub Issues](https://github.com/kirimku/smartseller-backend/issues)
- **Email**: support@smartseller.com

---

**SmartSeller** - Empowering every seller to succeed in the digital marketplace 🚀
