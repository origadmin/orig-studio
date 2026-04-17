# OrigCMS Architecture

## Table of Contents

- [1. Overview](#1-overview)
- [2. System Architecture](#2-system-architecture)
- [3. Service Architecture](#3-service-architecture)
- [4. Data Model](#4-data-model)
- [5. API Design](#5-api-design)
- [6. Technology Stack](#6-technology-stack)

---

## 1. Overview

OrigCMS is a high-performance open-source media content management system built with Go microservices architecture. It provides features for video hosting, video-on-demand, and live streaming, suitable for educational institutions, enterprise intranets, and community portals.

### Key Features

- **Go Microservices Architecture** — Based on go-kratos framework, high concurrency, low memory footprint
- **HLS Adaptive Streaming** — Automatic multi-resolution transcoding with seamless quality switching
- **Multiple Media Types** — Support for video, audio, image, and PDF
- **Flexible Publishing Workflow** — Public, private, and unlisted visibility controls
- **RBAC Role Permissions** — Fine-grained user group and permission management
- **Category and Tag System** — Multi-dimensional media organization
- **Playlists** — Organized and shared video/audio content
- **Comments and Interactions** — Like, favorite, and comment system
- **Responsive Design** — Adapts to desktop and mobile with light/dark theme support
- **i18n Internationalization** — Built-in multi-language support framework
- **REST API** — Complete API interface for integration and secondary development
- **JWT Authentication** — Secure token authentication mechanism

---

## 2. System Architecture

```
                                    ┌─────────────────────────────────────┐
                                    │           API Gateway               │
                                    │     (HTTP/gRPC Routing)            │
                                    └──────────────┬────────────────────┘
                                                   │
            ┌────────────────┬─────────────────────┼─────────────────────┬────────────────┐
            │                │                     │                     │                │
            ▼                ▼                     ▼                     ▼                ▼
    ┌──────────────┐ ┌──────────────┐     ┌──────────────┐     ┌──────────────┐ ┌──────────────┐
    │  svc-user    │ │  svc-media   │     │ svc-content  │     │  svc-admin   │ │ svc-system   │
    │  (User)      │ │  (Media)    │     │  (Content)  │     │  (Admin)     │ │  (Stats)     │
    └──────┬───────┘ └──────┬───────┘     └──────┬───────┘     └──────────────┘ └──────────────┘
           │                │                     │
           ▼                ▼                     ▼
    ┌──────────────────────────────────────────────────────────────┐
    │                    PostgreSQL Database                         │
    │                  (via Ent ORM)                                │
    └──────────────────────────────────────────────────────────────┘
           │                │                     │
           ▼                ▼                     ▼
    ┌──────────────┐ ┌──────────────┐     ┌──────────────┐
    │    Redis     │ │   FFmpeg     │     │    Redis     │
    │   (Cache)    │ │  (Transcode) │     │  (Message)   │
    └──────────────┘ └──────────────┘     └──────────────┘
```

---

## 3. Service Architecture

### 3.1 Service Layer (biz)

Business logic layer implementing use cases and domain models.

| Service | Module | Responsibility |
|---------|--------|----------------|
| **svc-user** | UserUseCase | User CRUD, authentication, password management |
| **svc-user** | Subscription | Channel subscription management |
| **svc-media** | MediaUseCase | Media CRUD, upload, transcoding |
| **svc-media** | TranscodeWorker | Video encoding task processing |
| **svc-content** | CommentUseCase | Comment management |
| **svc-content** | FeedUseCase | Home feed, trending feed |
| **svc-content** | LikeFavoriteUseCase | Like and favorite management |
| **svc-content** | NotificationUseCase | User notifications |
| **svc-admin** | TagUseCase | Tag management |

### 3.2 Data Access Layer (data)

Repository pattern implementations for database operations.

```
data/
├── entity/           # Ent ORM generated files
├── enums/           # Enumeration definitions
└── {module}/        # Repository implementations
    ├── user_repo.go
    ├── media_repo.go
    └── comment_repo.go
```

### 3.3 Service Layer (service)

gRPC/HTTP handlers that expose business operations.

```
service/
├── user.go          # UserService handlers
├── media.go        # MediaService handlers
└── upload.go       # Upload handlers
```

---

## 4. Data Model

### 4.1 Entity Relationship

```
┌─────────┐       ┌─────────────┐       ┌─────────┐
│  User   │──────│  Channel    │───────│Subscription│
└────┬────┘       └──────┬──────┘       └──────────┘
     │                    │
     │                    │
     ▼                    ▼
┌─────────┐       ┌─────────────┐       ┌─────────────┐
│  Media  │──────│   Comment   │       │    Like     │
└────┬────┘       └──────┬──────┘       └──────┬──────┘
     │                    │                    │
     │                    ▼                    ▼
     │             ┌─────────────┐       ┌─────────────┐
     │             │  Favorite   │       │  Playlist   │
     │             └─────────────┘       └──────┬──────┘
     │                                         │
     ▼                                         │
┌─────────────┐                                │
│EncodingTask │                                │
└─────────────┘                                │
     │                                         │
     ▼                                         ▼
┌─────────────────┐                     ┌─────────────┐
│ EncodeProfile   │                     │MediaPlaylist│
└─────────────────┘                     └─────────────┘
```

### 4.2 Core Entities

| Entity | Description | Key Fields |
|--------|-------------|------------|
| **User** | System user | ID, Username, Email, Password, Role, Status |
| **Media** | Media content | ID, Title, URL, MimeType, Duration, UserID, CategoryID |
| **Channel** | User channel | ID, UserID, Title, Slug, Description |
| **Category** | Media category | ID, Title, Slug, ParentID |
| **Tag** | Media tag | ID, Title, Slug |
| **Comment** | Media comment | ID, MediaID, UserID, Text, ParentID, Status |
| **Like** | Media like | ID, MediaID, UserID, Type |
| **Favorite** | Media favorite | ID, MediaID, UserID |
| **Playlist** | Media playlist | ID, UserID, Title, IsPublic |
| **EncodingTask** | Transcoding task | ID, MediaID, ProfileID, Status, OutputPath |
| **EncodeProfile** | Encoding preset | ID, Name, Resolution, VideoCodec, Bitrate |
| **Subscription** | Channel subscription | ID, SubscriberID, ChannelID |
| **Notification** | User notification | ID, UserID, Type, Content, Read |

---

## 5. API Design

### 5.1 API Protocol

- **Primary**: gRPC with Protocol Buffers
- **Secondary**: HTTP/REST via gRPC-Gateway
- **API Versioning**: URL path `/api/v1/`

### 5.2 Service Endpoint Summary

| Service | gRPC Port | HTTP Endpoint Prefix | Description |
|---------|-----------|---------------------|-------------|
| **UserService** | 9000 | /api/v1/auth, /api/v1/users | User authentication and management |
| **MediaService** | 9000 | /api/v1/medias | Media CRUD and streaming |
| **CategoryService** | 9000 | /api/v1/categories | Category management |
| **TagService** | 9000 | /api/v1/tags | Tag management |
| **CommentService** | 9000 | /api/v1/comments | Comment management |
| **PlaylistService** | 9000 | /api/v1/playlists | Playlist management |
| **ChannelService** | 9000 | /api/v1/channels | Channel management |
| **SearchService** | 9000 | /api/v1/search | Search functionality |
| **AdminService** | 9000 | /api/v1/admin | Admin dashboard and settings |
| **PortalService** | 9000 | /api/v1/portal | Portal content delivery |

### 5.3 Authentication

JWT-based authentication with access and refresh tokens.

```
Login: POST /api/v1/auth/login
  Request: { "username": "xxx", "password": "xxx" }
  Response: { "access_token": "xxx", "refresh_token": "xxx", "expires_at": 1234567890 }

Headers: Authorization: Bearer <access_token>
```

---

## 6. Technology Stack

### Backend

| Component | Technology | Version |
|-----------|------------|---------|
| Language | Go | 1.21+ |
| Framework | go-kratos | v2 |
| Web Framework | Gin | Latest |
| ORM | ent | v0.14.6 |
| Database | PostgreSQL | 15+ |
| Cache | Redis | Latest |
| Message Queue | Watermill | Latest |
| Service Discovery | Consul | Latest |
| API Protocol | gRPC + REST | - |
| DI | Wire | Latest |
| Video Transcoding | FFmpeg | Latest |

### Frontend

| Component | Technology | Version |
|-----------|------------|---------|
| Framework | React | 18+ |
| Language | TypeScript | 5+ |
| UI Library | shadcn/ui | Latest |
| Build Tool | Rsbuild | Latest |
| Router | TanStack Router | Latest |
| State | TanStack Query | Latest |
| i18n | Built-in | - |

### DevOps

| Component | Technology |
|-----------|------------|
| Container | Docker |
| Orchestration | docker-compose |
| CI/CD | GitHub Actions |
| Deployment | Single Server / Microservices |

---

## 7. Directory Structure

```
orig-cms/
├── api/                          # API definitions
│   ├── proto/v1/                 # Protobuf definitions
│   │   ├── user/                # User service proto
│   │   ├── media/               # Media service proto
│   │   └── types/               # Shared types
│   └── gen/                     # Generated Go code
│
├── cmd/                          # Entry points
│   └── server/                  # Monolithic mode server
│
├── internal/                      # Internal packages
│   ├── auth/                    # JWT authentication
│   ├── conf/                    # Configuration
│   ├── data/                    # Data layer
│   │   ├── entity/              # Ent ORM schemas
│   │   ├── conv/                # Converters
│   │   └── enums/               # Enumerations
│   ├── helpers/                 # Helper utilities
│   ├── pubsub/                  # Pub/Sub messaging
│   ├── server/                  # HTTP/gRPC server
│   ├── svc-user/                # User service
│   ├── svc-media/               # Media service
│   ├── svc-content/             # Content service
│   ├── svc-admin/               # Admin service
│   └── svc-system/              # System service
│
├── configs/                      # Configuration files
├── docs/                        # Documentation
├── web/                         # Frontend (React)
└── resources/                   # Static resources
```

---

## 8. Deployment Modes

### 8.1 Monolithic Mode (Development)

Single process containing all services.

```
go run ./cmd/server/...
```

### 8.2 Microservices Mode (Production)

Independent services with service discovery.

```
┌─────────────┐
│   Gateway   │
└──────┬──────┘
       │
   ┌───┴───┬───────┬───────┐
   ▼       ▼       ▼       ▼
svc-user svc-media svc-content svc-admin
```

---

## 9. Security

### 9.1 Authentication Flow

```
User Login
    │
    ▼
┌─────────┐    JWT    ┌─────────┐
│ Client  │──────────▶│ Server  │
└─────────┘           └────┬────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ Validate    │
                    │ Credentials │
                    └──────┬──────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ Return JWT   │
                    │ Token        │
                    └─────────────┘
```

### 9.2 Authorization

Role-Based Access Control (RBAC) with the following roles:

| Role | Description |
|------|-------------|
| **admin** | Full system access |
| **editor** | Can manage media and comments |
| **user** | Can upload media, comment |
| **guest** | Can view public content |

---

*Document Version: 1.0*
*Last Updated: 2026-04-17*