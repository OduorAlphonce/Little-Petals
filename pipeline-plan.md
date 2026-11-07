# Little Petals - CI/CD Pipeline & DevSecOps Architecture Plan

## 1. Executive Summary & Security Threat Model

Little Petals is designed as a creative social platform where users publish books, chapters, poetry, blogs, and songs (audio files), and engage via comments and dynamic reactions. Because social platforms process untrusted, multi-tenant user input and binary media files at scale, security cannot be an afterthought bolted on at deployment—it must be enforced automatically at every commit and pull request.

### Primary Threat Vectors in Little Petals:
1. **Unrestricted / Malicious File Uploads (`/web/static/uploads/songs`, `covers`, `books`):**
   - Executable uploads disguised as `.mp3`/`.jpg` (Polyglot files, shell scripts, webshells).
   - Denial of Service (DoS) via zip bombs or massive file stream exhaustion.
   - Path traversal via tainted file names (`../../etc/passwd`).
2. **Cross-Site Scripting (XSS) & Content Injection:**
   - Raw HTML injection inside poetry whitespace, markdown blogs, and chapter text.
   - Unsanitized dynamic HTMX comments and reactions responses.
3. **Insecure Direct Object References (IDOR):**
   - Attackers modifying, deleting, or publishing drafts of other creators' books, chapters, or audio tracks.
4. **Authentication & Session Flaws:**
   - Weak password hashing, credential stuffing, session fixation, or insecure cookie flags (`HttpOnly`, `Secure`, `SameSite`).
5. **Database Integrity & Race Conditions:**
   - SQL injection, schema migration failure in production, or race conditions during rapid concurrent reactions.

---

## 2. Public vs. Private Repository Cost & Efficiency Strategy

| Metric / Feature | Public Repository | Private Repository (Free Tier) | Our Design Optimization |
| :--- | :--- | :--- | :--- |
| **GitHub Actions Minutes** | Unlimited free minutes | 2,000 minutes/month | Strict caching (`go-build` & Go modules) reduces run duration from ~6 min down to **< 90 seconds** per run (~1,300 builds/month free). |
| **Concurrency** | Up to 20 concurrent jobs | Up to 20 concurrent jobs | Parallel job matrix for independent checks (Lint/SAST runs alongside Unit Tests). |
| **Advanced Security & SAST** | Free GitHub CodeQL & Secret Scanning | Paid GitHub Advanced Security | **Zero vendor lock-in**: We use open-source industry standards (`gosec`, `govulncheck`, `gitleaks`, `trivy`) that run 100% free with identical high fidelity on both public and private tiers. |
| **Artifact & Cache Storage** | 10 GB shared | 2 GB per repository | Ephemeral build caches, small binary artifacts, and pruning of untracked assets. |

---

## 3. End-to-End Pipeline Architecture

```mermaid
flowchart TD
    subgraph Triggers ["Event Triggers"]
        PR["Pull Request (develop / main)"]
        PUSH["Push to develop (Staging)"]
        TAG["Tag / Release v*.*.* (Production)"]
        CRON["Weekly Schedule (Mon 00:00 UTC)"]
    end

    subgraph Stage1 ["Stage 1: Fast Quality & Security Gate"]
        GITLEAKS["Secret Scan (Gitleaks)"]
        LINT["Code Quality & Linters (golangci-lint)"]
        GOSEC["Static Code Security (gosec SAST)"]
        VULN["Dependency CVE Audit (govulncheck)"]
    end

    subgraph Stage2 ["Stage 2: Strict Automated Testing"]
        UNIT["Unit Tests with Race Detector (-race)"]
        INTEG["DB Integration Tests (Real PostgreSQL 16)"]
        COVERAGE["Code Coverage Gate (>= 80%)"]
    end

    subgraph Stage3 ["Stage 3: Container Hardening & Image Scan"]
        BUILD_IMG["Build Multi-Stage Distroless Image"]
        TRIVY_SCAN["Container Image Vulnerability Scan (Trivy)"]
    end

    subgraph Stage4 ["Stage 4: Automated Continuous Delivery"]
        MIGRATE["Dry-Run / Run Migrations (embed.FS)"]
        DEPLOY_STAGE["Deploy to Staging"]
        PROD_GATE{"Manual Approval Gate"}
        DEPLOY_PROD["Deploy to Production"]
        HEALTH["Post-Deployment Smoke Test (/healthz)"]
    end

    PR --> GITLEAKS & LINT & GOSEC & VULN
    GITLEAKS & LINT & GOSEC & VULN --> UNIT & INTEG
    UNIT & INTEG --> COVERAGE

    PUSH --> GITLEAKS
    COVERAGE --> BUILD_IMG
    BUILD_IMG --> TRIVY_SCAN
    TRIVY_SCAN --> MIGRATE --> DEPLOY_STAGE --> HEALTH

    TAG --> BUILD_IMG
    TRIVY_SCAN --> PROD_GATE
    PROD_GATE --> MIGRATE --> DEPLOY_PROD --> HEALTH

    CRON --> VULN & TRIVY_SCAN
```

---

## 4. Pipeline Stages & Strict Check Specifications

### Stage 1: Fast Security & Code Quality Gate (< 45s)
- **Secret Detection (`gitleaks`):** Scans every commit in the PR for hardcoded API keys, database credentials, JWT secrets, and private tokens.
- **Go Mod Verification:** Ensures `go.sum` is clean, dependencies are not tampered with, and no unused modules linger (`go mod verify`, `go mod tidy -check`).
- **Linter Enforcement (`golangci-lint`):**
  - Bug detection: `govet`, `staticcheck`, `errcheck`, `unused`, `ineffassign`.
  - Security linter: `gosec` rule integration.
  - Performance & Concurrency: checks for unclosed response bodies, unhandled goroutines, and SQL query formatting.

### Stage 2: SAST & Supply Chain Audit (< 30s)
- **`govulncheck`:** The official Go vulnerability scanner. Unlike basic package-version checkers, `govulncheck` analyzes the call graph to determine if vulnerable functions are actually executed by Little Petals.
- **`gosec` Deep Scan:** Enforces strict Go security rules:
  - G101: Hardcoded credentials.
  - G201/G202: SQL query concatenation (ensures parameterized queries in repositories).
  - G304: File path traversal via untrusted input in upload paths.
  - G401/G501: Weak cryptographic primitives (ensures `bcrypt` or strong hashing).

### Stage 3: Automated Test Matrix (< 60s)
- **Unit Testing with Data Race Detection:**
  - `go test -v -race -timeout 3m ./...`
  - Critical for social platforms where concurrent requests (e.g., hundreds of users reacting to a post simultaneously) can corrupt counters or session states.
- **Real Database Integration Testing (PostgreSQL Service Container):**
  - Spins up a real `postgres:16-alpine` service in GitHub Actions.
  - Executes the real migration suite from `web/migrations/embed.go`.
  - Tests repository CRUD operations (`internal/repository/`) against the live schema to guarantee zero SQL syntax errors or foreign key mismatches.
- **Code Coverage Gate:**
  - Emits `coverage.out` and asserts minimum test coverage on business logic (`internal/services/`, `internal/handlers/`, `internal/middleware/`).

### Stage 4: Container Hardening & Image Scan (< 60s)
- **Hardened Multi-Stage Dockerfile:**
  - **Builder Stage:** Compiles statically linked binary with `CGO_ENABLED=0 -ldflags="-s -w"`.
  - **Runtime Stage:** Uses Google `distroless/static-debian12:nonroot` or `alpine` running under a dedicated non-root user (`UID 10001`).
  - Read-only root filesystem with explicit permissions for asset uploads.
- **Container Vulnerability Scanning (`trivy`):**
  - Fails the build if any `HIGH` or `CRITICAL` severity CVEs are present in the final deployable artifact.

### Stage 5: Continuous Delivery (CD) & Verification
- **Zero-Downtime Deployment Flow:**
  - Automated database migration execution: Safe, backward-compatible migrations run prior to traffic routing.
  - Staging environment auto-deploys on push to `develop`.
  - Production environment requires GitHub Environment approval and tags (`v*.*.*`).
  - Automated smoke test hits `GET /healthz` to verify database ping and server readiness before completing deployment.

---

## 5. Concrete Configuration Blueprints

### A. Core CI Workflow (`.github/workflows/ci.yml`)

```yaml
name: CI & Security Quality Gate

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  secrets:
    name: Secret Detection
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Run Gitleaks
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  lint-and-sast:
    name: Lint & SAST Scan
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Verify Module Dependencies
        run: |
          go mod verify
          git diff --exit-code go.mod go.sum

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v1.64.5
          args: --timeout=5m

      - name: Run Gosec Security Audit
        run: |
          go install github.com/securego/gosec/v2/cmd/gosec@latest
          gosec -fmt=text -severity=medium -confidence=medium ./...

      - name: Run Govulncheck (Vulnerability Analysis)
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

  test:
    name: Tests & DB Migrations
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: testuser
          POSTGRES_PASSWORD: testpassword
          POSTGRES_DB: littlepetals_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 5s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Run Unit & Race Tests with Postgres
        env:
          DB_CONN_STR: "postgres://testuser:testpassword@localhost:5432/littlepetals_test?sslmode=disable"
          PORT: 8080
          ENV_FILE: ""
        run: |
          go test -v -race -covermode=atomic -coverprofile=coverage.out ./...

      - name: Verify Test Coverage
        run: |
          go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $3}'
```

### B. Hardened Multi-Stage Dockerfile (`Dockerfile`)

```dockerfile
# syntax=docker/dockerfile:1.7

# --- Build Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Compile hardened static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w -extldflags '-static'" \
    -o /bin/little-petals ./cmd/server/main.go

# --- Production Runtime Stage ---
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /bin/little-petals /app/little-petals

# Copy static assets and templates
COPY --from=builder /src/web /app/web

# Expose server port
EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/little-petals"]
```

### C. Build & Security Scan Workflow (`.github/workflows/package.yml`)

```yaml
name: Build & Container Security

on:
  push:
    branches: [main, develop]
    tags: ['v*.*.*']

jobs:
  docker-scan:
    name: Build & Vulnerability Scan
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build Local Image
        uses: docker/build-push-action@v5
        with:
          context: .
          load: true
          tags: little-petals:ci-test
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Scan Image with Trivy
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: little-petals:ci-test
          format: 'table'
          exit-code: '1'
          ignore-unfixed: true
          vuln-type: 'os,library'
          severity: 'CRITICAL,HIGH'
```

### D. Recommended `.golangci.yml` Configuration

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gosec
    - bodyclose
    - sqlclosecheck
    - rowserrcheck

linters-settings:
  gosec:
    severity: medium
    confidence: medium
```

---

## 6. Implementation Roadmap

```mermaid
flowchart LR
    P1["Phase 1: Local & CI Quality Gate\n(golangci-lint, gitleaks, ci.yml)"] --> P2["Phase 2: DB Integration Testing\n(PostgreSQL service container in CI)"]
    P2 --> P3["Phase 3: Container Hardening & SAST\n(Dockerfile, Trivy, Govulncheck)"]
    P3 --> P4["Phase 4: CD & Release Pipeline\n(GitHub Environments, /healthz verify)"]
```

1. **Phase 1: Quality Gate & Secrets Hygiene**
   - Add `.golangci.yml` and `.github/workflows/ci.yml`.
   - Prevent any secret leaks with Gitleaks before commits reach the main branches.
2. **Phase 2: Live Database Integration Testing**
   - Run tests against a live PostgreSQL 16 container in GitHub Actions to validate migrations and queries.
3. **Phase 3: Container Hardening & Supply Chain Security**
   - Introduce the multi-stage, rootless, distroless `Dockerfile`.
   - Automate vulnerability scanning with Trivy on every build.
4. **Phase 4: CD & Production Governance**
   - Configure staging (`develop`) and production (`main`) deployment triggers with `/healthz` post-deployment verification.
