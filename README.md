# SOA 2025 Backend - Microservices Architecture

A microservices-based backend system built with Go and Node.js for a tour management platform.

## Overview

This system provides a complete backend solution for managing tours, users, social connections, and purchases. It follows a microservices architecture where each service handles a specific business domain and communicates with others through well-defined APIs.

### What it does

- **User Management**: Handles user registration, authentication, and profile management with role-based access (Guides, Tourists, Admins)
- **Tour Management**: Allows guides to create and manage tours with keypoints, while tourists can browse, search, and view tour details
- **Social Features**: Enables users to follow each other and build a social network around shared interests
- **Blog System**: Provides a platform for users to share experiences, write blog posts, and interact through comments and likes
- **Purchases**: Manages shopping cart functionality and handles tour purchases with token-based access control
- **Content Management**: Handles image uploads for profiles, tour reviews, and tour keypoints
- **Location Services**: Integrates with map services to provide location-based features

### How it works

The system is accessed through a single **API Gateway** that routes requests to the appropriate microservice. Each service is independent and owns its data:

- When a user logs in, the Gateway forwards the request to the **Stakeholders Service** which authenticates and returns a JWT token
- To browse tours, the Gateway routes to the **Tours Service** which queries MongoDB for tour data
- When purchasing a tour, the **Purchase Service** coordinates with Tours and Stakeholders services to complete the transaction
- Social connections are managed by the **Follower Service** using a graph database to efficiently query relationships
- Images are stored and served through the **Image Service** which handles uploads and retrieval

All services communicate through REST APIs, with some using gRPC for high-performance inter-service communication. Each service can be developed, deployed, and scaled independently.

## Architecture

```
Gateway (8080) → Stakeholders Service → PostgreSQL
              → Tours Service → MongoDB
              → Blog Service → PostgreSQL (Prisma)
              → Purchase Service → PostgreSQL
              → Follower Service → Neo4j
              → Image Service
              → Map Service
```

## Services

| Service | Technology | Port | Database |
|---------|-----------|------|----------|
| **Gateway** | Go (Gin) | 8080 | - |
| **Stakeholders** | Go | 8081 | PostgreSQL (5433) |
| **Tours** | Go | 8082 | MongoDB (5435) |
| **Blog** | Node.js (Express) | 3000 | PostgreSQL (5434) |
| **Purchase** | Go | 8084 | PostgreSQL (5436) |
| **Follower** | Go | 8083 | Neo4j (7687) |
| **Image** | Node.js (Express) | 3001 | File Storage |
| **Map** | Node.js (Express) | 3002 | External API |

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.23+ (for local development)
- Node.js 18+ (for local development)

### Run with Docker Compose

1. **Clone repository**
   ```bash
   git clone <repository-url>
   cd soa-2025-be
   ```

2. **Set up environment variables**
   
   Create `.env` files for each service (see Configuration section).

3. **Start all services**
   ```bash
   docker-compose up -d
   ```

4. **Check status**
   ```bash
   docker-compose ps
   docker-compose logs -f
   ```

5. **Access services**
   - Gateway: http://localhost:8080
   - Individual services on their respective ports (see table above)

## Configuration

### Required Environment Variables

#### Gateway
```env
GATEWAY_PORT=8080
JWT_SECRET=<base64-encoded-secret>
BLOG_SERVICE_URL=http://blog-service:3000
IMAGE_SERVICE_URL=http://image-service:3000
STAKEHOLDERS_SERVICE_URL=http://stakeholders-service:8080
FOLLOWER_SERVICE_URL=http://follower-service:8080
TOURS_SERVICE_GRPC_URL=tours-service:50051
TOURS_SERVICE_API_URL=http://tours-service:8080
PURCHASE_SERVICE_URL=http://purchase-service:8080
```

#### Stakeholders Service
```env
DB_HOST=stakeholders-db
DB_PORT=5432
DB_USER=<username>
DB_PASSWORD=<password>
DB_NAME=<database-name>
JWT_SECRET=<base64-encoded-secret>
JWT_EXPIRATION=60
IMAGE_SERVICE_URL=http://image-service:3000
```

#### Tours Service
```env
DB_HOST=tours-db
DB_PORT=27017
DB_USER=<username>
DB_PASSWORD=<password>
DB_NAME=<database-name>
AUTH_DB=admin
MAP_SERVICE_URL=http://map-service:3000
STAKEHOLDERS_SERVICE_URL=http://stakeholders-service:8080
```

#### Purchase Service
```env
DB_HOST=purchase-db
DB_PORT=5432
DB_USER=<username>
DB_PASSWORD=<password>
DB_NAME=<database-name>
STAKEHOLDERS_SERVICE_URL=http://stakeholders-service:8080
TOURS_SERVICE_URL=http://tours-service:8080
```

#### Follower Service
```env
PORT=8080
NEO4J_URI=bolt://neo4j:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=<password>
```

#### Blog Service
```env
PORT=3000
DATABASE_URL=postgresql://postgres:password@blog-db:5432/blogSOA?schema=public
```

### Generate JWT Secret
```bash
openssl rand -base64 32
```

## API Overview

### Authentication
Most endpoints require JWT token in header:
```
Authorization: Bearer <token>
```

## Development

### Local Development

#### Go Services
```bash
cd {service-name}
go mod download
go run cmd/server/main.go
```

#### Node.js Services
```bash
cd {service-name}
npm install
npm start
```

#### Blog Service (Prisma)
```bash
cd blog-service
npx prisma generate
npx prisma migrate deploy
npm start
```

👥 Authors
Duško Pilipović, Ognjen Papović, Nemanja Zekanović, Nikola Pejanović
