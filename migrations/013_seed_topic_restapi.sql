-- ============================================================================
-- Seed: Topic-only challenge — RESTful API Design (challenge-2)
-- ============================================================================

INSERT INTO challenges (
    id, title, description, difficulty, type, max_score, 
    repo_template_url, requirements, tech_stack, estimated_hours, 
    is_published, created_at, updated_at, description_md, rubric
) VALUES (
    '1',
    'Understanding REST APIs',
    'Learn about RESTful architecture, HTTP methods, status codes, and API design best practices.',
    'Easy',
    'topic',
    100,
    NULL,
    '["Understand HTTP methods (GET, POST, PUT, DELETE)", "Learn about status codes", "Design a RESTful endpoint structure"]',
    '["REST", "HTTP", "API Design"]',
    2,
    TRUE,
    NOW(),
    NOW(),
    $$## What is REST?

REST (Representational State Transfer) is an **architectural style** for designing networked applications. It was introduced by Roy Fielding in his 2000 doctoral dissertation and has since become the dominant standard for building web APIs.

REST relies on a **stateless, client-server communication protocol** — almost always HTTP.

---

## Core Principles

### 1. Client-Server Separation
The client (frontend) and server (backend) are independent. The server doesn't know about the UI, and the client doesn't know about data storage. They communicate through a well-defined API contract.

### 2. Statelessness
Each request from a client must contain **all the information** needed to process it. The server does not store any session state between requests.

```
GET /api/users/42
Authorization: Bearer eyJhbGciOiJSUzI1NiIs...
```

> **Why it matters:** Statelessness makes your API horizontally scalable — any server instance can handle any request.

### 3. Uniform Interface
Resources are identified by **URIs**, manipulated through **representations** (JSON, XML), and interactions are driven by **standard HTTP methods**.

### 4. Cacheability
Responses must define whether they are cacheable or not. Proper caching reduces server load and improves perceived performance.

### 5. Layered System
A client cannot tell whether it is connected directly to the server or through intermediaries (load balancers, CDNs, API gateways).

---

## HTTP Methods

HTTP defines a set of request methods (verbs) that indicate the desired action on a resource:

| Method | Purpose | Idempotent | Safe |
|--------|---------|------------|------|
| `GET` | Retrieve a resource | ✅ | ✅ |
| `POST` | Create a new resource | ❌ | ❌ |
| `PUT` | Replace an entire resource | ✅ | ❌ |
| `PATCH` | Partially update a resource | ❌ | ❌ |
| `DELETE` | Remove a resource | ✅ | ❌ |

### Examples

```http
GET    /api/v1/users          → List all users
GET    /api/v1/users/42       → Get user with ID 42
POST   /api/v1/users          → Create a new user
PUT    /api/v1/users/42       → Replace user 42 entirely
PATCH  /api/v1/users/42       → Update specific fields of user 42
DELETE /api/v1/users/42       → Delete user 42
```

> **Idempotent** means calling the same request multiple times produces the same result. `GET`, `PUT`, and `DELETE` are idempotent. `POST` is not — calling it twice creates two resources.

---

## HTTP Status Codes

Status codes tell the client what happened with their request:

### 2xx — Success
- **200 OK** — Request succeeded (general success)
- **201 Created** — Resource was successfully created
- **204 No Content** — Success, but no body to return (common for DELETE)

### 3xx — Redirection
- **301 Moved Permanently** — Resource has a new permanent URI
- **304 Not Modified** — Cached version is still valid

### 4xx — Client Errors
- **400 Bad Request** — Malformed request or validation failure
- **401 Unauthorized** — Authentication required or failed
- **403 Forbidden** — Authenticated but not authorized
- **404 Not Found** — Resource does not exist
- **409 Conflict** — Request conflicts with current state
- **422 Unprocessable Entity** — Validation errors
- **429 Too Many Requests** — Rate limit exceeded

### 5xx — Server Errors
- **500 Internal Server Error** — Something broke on the server
- **502 Bad Gateway** — Upstream server returned an invalid response
- **503 Service Unavailable** — Server is temporarily down

---

## URL Design Best Practices

### Use Nouns, Not Verbs
```
✅  GET  /api/v1/articles
❌  GET  /api/v1/getArticles
```

### Use Plural Nouns
```
✅  /api/v1/users
❌  /api/v1/user
```

### Use Nesting for Relationships
```
GET /api/v1/users/42/posts        → All posts by user 42
GET /api/v1/users/42/posts/7      → Post 7 by user 42
```

### Version Your API
```
/api/v1/users
/api/v2/users
```

### Use Query Parameters for Filtering, Sorting & Pagination
```
GET /api/v1/articles?status=published&sort=created_at&page=2&limit=20
```

---

## Request & Response Patterns

### Successful Response
```json
{
  "data": {
    "id": 42,
    "name": "Jane Doe",
    "email": "jane@example.com"
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Email is required",
    "details": [
      { "field": "email", "message": "must not be empty" }
    ]
  }
}
```

### Paginated List Response
```json
{
  "data": [...],
  "pagination": {
    "page": 2,
    "limit": 20,
    "total": 156,
    "has_next": true
  }
}
```

---

## Authentication Approaches

| Method | Use Case |
|--------|----------|
| **API Keys** | Simple server-to-server auth |
| **JWT (Bearer Tokens)** | Stateless user authentication |
| **OAuth 2.0** | Third-party delegated access |
| **Session Cookies** | Traditional web apps |

```http
GET /api/v1/me
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## Your Task

Design a RESTful API for a **Book Library Management System** with the following resources:

1. **Books** — CRUD operations with search and filtering
2. **Authors** — Linked to books (one author → many books)
3. **Reviews** — Users can review books (nested under books)
4. **Users** — Authentication and profile management

### Requirements

- Design the complete URL structure for all endpoints
- Choose appropriate HTTP methods and status codes
- Define request/response JSON schemas for at least 3 endpoints
- Include pagination, filtering, and error handling patterns
- Document your API in a README.md file

### Bonus Points

- Implement the API using any backend framework (Express, Gin, Django, etc.)
- Add rate limiting and input validation
- Include OpenAPI/Swagger documentation
$$,
    '{}'::jsonb
)
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    difficulty = EXCLUDED.difficulty,
    type = EXCLUDED.type,
    max_score = EXCLUDED.max_score,
    repo_template_url = EXCLUDED.repo_template_url,
    requirements = EXCLUDED.requirements,
    tech_stack = EXCLUDED.tech_stack,
    estimated_hours = EXCLUDED.estimated_hours,
    is_published = EXCLUDED.is_published,
    description_md = EXCLUDED.description_md,
    rubric = EXCLUDED.rubric,
    updated_at = NOW();

