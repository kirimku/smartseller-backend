# Kirimku Backend

A robust Go backend service built using clean architecture principles, providing secure and scalable API endpoints.

[![Build and Deploy](https://github.com/kirimku/kirimku-backend/actions/workflows/main.yml/badge.svg)](https://github.com/kirimku/kirimku-backend/actions/workflows/main.yml)
[![Deployment](https://img.shields.io/badge/deployment-DigitalOcean%20App%20Platform-blue)](https://docs.digitalocean.com/products/app-platform/)
[![Container Registry](https://img.shields.io/badge/registry-GitHub%20Container%20Registry-brightgreen)](https://github.com/kirimku/kirimku-backend/pkgs/container/kirimku-backend)

## Features

- 🔒 Secure Authentication with Google OAuth
- 🚦 Rate Limiting and CORS Protection
- 📝 OpenAPI/Swagger Documentation
- 🧪 Comprehensive Test Coverage
- 🗄️ Database Migrations
- 🔍 Input Validation and Sanitization
- 🎯 Clean Architecture Implementation

## Architecture

This project follows clean architecture principles with the following layers:

- **Domain Layer** (`internal/domain/`)
  - Contains business logic and entities
  - Defines repository and usecase interfaces
  - Independent of external frameworks

- **Use Case Layer** (`internal/usecase/`)
  - Implements business logic
  - Orchestrates data flow between layers
  - Contains application-specific business rules

- **Interface Layer** (`internal/interfaces/`)
  - HTTP handlers and middleware
  - Request/Response handling
  - Input validation

- **Infrastructure Layer** (`internal/infrastructure/`)
  - Database implementations
  - External service integrations
  - Repository implementations

## Prerequisites

- Go 1.19 or higher
- PostgreSQL 13 or higher
- Make

## Environment Setup

1. Clone the repository:
```bash
git clone git@github.com:kirimku/kirimku-backend.git
cd kirimku-backend
```

2. Create a .env file with required configurations:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=kirimku
GOOGLE_CLIENT_ID=your_client_id
GOOGLE_CLIENT_SECRET=your_client_secret
```

3. Install dependencies:
```bash
go mod tidy
```

## Development

### Running the Application

```bash
# Run the application
make run

# Run in development mode with hot reload
make dev

# Build the application
make build
```

### Database Migrations

```bash
# Run migrations up
make migrate-up

# Run migrations down
make migrate-down

# Create a new migration
make migrate-create name=migration_name
```

#### DigitalOcean Database Migrations

The project includes multiple ways to run database migrations on your DigitalOcean managed database:

1. **Automatic Migrations**: Migrations run automatically when:
   - Changes are pushed to migration files in the `internal/infrastructure/database/migrations/` or `migrations/` directories
   - The application is deployed via the deploy-digitalocean.yml workflow

2. **Manual Migrations via GitHub Actions**:
   ```bash
   # Run all migrations on DigitalOcean database via GitHub Actions
   ./scripts/do_migrate_db.sh
   ```

3. **SSH Tunnel Migrations**: For developers without static IPs, use the SSH tunnel method:
   ```bash
   # All-in-one command (powers on Droplet, runs migrations, powers off Droplet)
   make tunnel-migrate-up
   
   # For finer control:
   make tunnel-start          # Start SSH tunnel only
   make tunnel-migrate-status # Check migration status
   make tunnel-stop           # Stop SSH tunnel
   
   # Control Droplet power state (to minimize costs)
   make droplet-on
   make droplet-off
   ```

4. **CI/CD Pipeline**: The database migration workflow (`db-migration.yml`) is triggered:
   - On push to the main branch with changes to migration files 
   - After successful deployment via the deploy-digitalocean.yml workflow
   - Manually through GitHub Actions dispatch

For detailed information about SSH tunnel migrations, see the [SSH Tunnel Guide](./docs/SSH_TUNNEL_GUIDE.md).
For CI/CD pipeline information, see the `.github/workflows` directory.

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# View coverage report in browser
make coverage-html
```

#### Admin Endpoint Testing

For testing admin functionality, use the predefined test admin account:

```bash
# Set up test admin user (automated)
./scripts/setup_test_admin.sh

# Quick admin endpoint test
./quick_test_admin_users.sh

# Comprehensive admin endpoint test suite
./test_admin_users_endpoint.sh
```

**Test Admin Credentials:**
- Email: `admin.test@kirimku.com`  
- Password: `TestAdmin123!`

See [Test Admin Credentials](./docs/TEST_ADMIN_CREDENTIALS.md) for complete setup instructions.

## API Documentation

The API is documented using OpenAPI/Swagger. After starting the server, visit:
- Swagger UI: `http://localhost:8080/swagger/`
- OpenAPI JSON: `http://localhost:8080/swagger/doc.json`

### Key Endpoints

- **Authentication**
  - `GET /auth/google/login` - Initiate Google OAuth login
  - `GET /auth/google/callback` - Handle OAuth callback
  - `POST /auth/refresh` - Refresh access token

## Security

The application implements several security measures:

- Rate limiting to prevent brute force attacks
- CORS configuration for API access control
- Input validation and sanitization
- Secure session management
- HTTP security headers

## Error Handling

The application uses structured error responses:

```json
{
  "status": "error",
  "message": "Detailed error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

## Monitoring and Logging

- Structured logging with log levels
- Request/Response logging
- Error tracking
- Performance metrics

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Deployment

### DigitalOcean App Platform

This project is configured for deployment on DigitalOcean App Platform with managed PostgreSQL database. We've optimized for cost by using the lowest tier specifications.

#### Automated Deployment with GitHub Actions

The recommended deployment method is using GitHub Actions, which automatically deploys your application whenever you push to the `main` branch:

1. Set up GitHub Actions secrets in your repository:
   - `DO_TOKEN`: Your DigitalOcean API token
   - `DO_APP_ID`: Your DigitalOcean App ID

2. Push your code to GitHub:
   ```bash
   git push origin main
   ```

3. Monitor the deployment in GitHub Actions tab of your repository

For detailed information about the GitHub Actions and Container Registry setup, see:
- [GitHub Actions & Container Registry Guide](./docs/GITHUB_ACTIONS_SETUP.md)

#### Manual Deployment

You can also deploy manually using the DigitalOcean CLI:

```bash
# Install doctl if not already installed
brew install doctl
doctl auth init

# Validate migrations before deployment
./scripts/validate_migrations.sh

# Deploy using the helper script
./scripts/digitalocean_deploy.sh
```

For detailed manual deployment instructions, see:
- [DigitalOcean Deployment Guide](./docs/DIGITALOCEAN_DEPLOYMENT.md)
- [Environment Variables Setup](./docs/DIGITALOCEAN_ENV_VARS.md)
- [Cost Estimation](./docs/DIGITALOCEAN_COST_ESTIMATION.md)

### Connecting to the Database

The DigitalOcean managed database requires trusted source IP addresses for access. For developers without static IPs, we provide an SSH tunnel solution:

1. **SSH Tunnel Setup**:
   - A small DigitalOcean Droplet serves as a tunnel to access the database
   - This approach is cost-efficient as the Droplet can be powered off when not in use
   - Makefile commands simplify tunnel management

2. **Available Commands**:
   ```bash
   # Database migrations via SSH tunnel
   make tunnel-migrate-up
   make tunnel-migrate-down
   make tunnel-migrate-status
   
   # Manual tunnel management
   make tunnel-start
   make tunnel-stop
   make tunnel-status
   
   # Droplet power management (to minimize costs)
   make droplet-on
   make droplet-off
   ```

3. **DBeaver or Other Tools**: You can connect your database tools through the SSH tunnel
   - See the detailed [SSH Tunnel Guide](./docs/SSH_TUNNEL_GUIDE.md) for instructions

For additional deployment options and configurations, refer to the documentation in the `docs` directory.