# OmniCam

This guide provides instructions for both developers and DevOps engineers working on the OmniCam project.

## 📋 Table of Contents

**<a id="development-guide">1. 🛠 Development Guide</a>**

- **1.1** <a href="#core-tools">Core Tools</a>
- **1.2** <a href="#frontend-development">Frontend Development</a>
  - **1.2.1** Setup & Installation
  - **1.2.2** Local Configuration
  - **1.2.3** Code Generation

- **1.3** <a href="#backend-development">Backend Development</a>
  - **1.3.1** <a href="#cli-utilities">CLI Utilities</a>
  - **1.3.2** <a href="#local-workflow">Local Workflow</a>
  - **1.3.3** <a href="#database--schema-management">Database & Schema Management</a>
    - **1.3.3.1** <a href="#migrations">Migrations</a>
    - **1.3.3.2** <a href="#queries">Queries</a>

- **1.4** <a href="#file-system-structure">File System Structure</a>
- **1.5** <a href="#linting-and-code-style">Linting and Code Style</a>
- **1.6** <a href="#recommended-extensions-vs-code">Recommended Extensions (VS Code)</a>

**2. <a href="#devops-deployment">🚢 DevOps & Deployment</a>**

- **2.1** <a href="#production-environment-setup">Production Environment Setup</a>

---

# [🛠 Development Guide]

## Core Tools

Ensure your local environment matches these global versions before proceeding:

- **Node.js**: v22.18.0
- **Golang**: v1.24.5
- **Docker**: For running PostgreSQL 17

## 🖥 Frontend Development

> **Note:** Add Go binaries to your path by adding `export PATH="$PATH:$(go env GOBIN)"` to your `~/.bashrc` or `~/.zshrc`.

**1. Setup & Installation**

```bash
cd frontend
npm install
```

**2. Local Configuration**
Create a `frontend/.env` file based on `.env.example`.

> **Note:** Never use `process.env` directly in components. Map values through the configuration schema.

**3. Code Generation**
Run these after any changes to SQL queries or Protobuf definitions:

```bash
npm run sqlc && npm run proto
```

## ⚙️ Backend Development

### CLI Utilities

Install these Go binaries for database management and Protobuf compilation.

1. **sqlc**: Generates type-safe Go code from SQL.

   ```bash
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   ```

2. **golang-migrate**: Handles database versioning.

   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```

3. **Protoc & Plugins**:
   - [Install Protoc Compiler](https://protobuf.dev/installation/)
   - **Golang Plugin**:
     ```bash
     go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
     ```

### Local Workflow

1. **Database Setup**: Start a PostgreSQL 17 instance via Docker:

   ```bash
   docker run --rm -it --name omnicam-postgres -e POSTGRES_PASSWORD=password -p 5433:5432 postgres:17
   ```

2. **Frontend Post-install**:

   ```bash
   npm install
   npm run sqlc && npm run proto
   ```

3. **Execution Commands**:

   | Task         | Command                    |
   | ------------ | -------------------------- |
   | **Frontend** | `npm run dev`              |
   | **Backend**  | `go run ./backend/main.go` |
   | **Swagger**  | `npm run swagger`          |

---

### Database & Schema Management

#### Migrations

Every schema change requires a .up.sql (apply) and a .down.sql (rollback) file.

- **Apply**:
  ```bash
  migrate -path db/migrations/ -database 'postgresql://postgres:password@localhost:5432/omnicam?sslmode=disable' up
  ```
- **Rollback**: Commands in the down file must be the exact reverse order of the up file.

#### Queries

We use **sqlc**. If you use the **Run on Save** VS Code extension, code generation will trigger automatically upon saving .sql files.

---

### File System Structure

The application expects the following directory structure for persistent storage:

- **3D Models**: `uploads/{projectId}/{id}/modelName`
- **Model Images**: `uploads/model/{projectId}/{id}/image.jpg`
- **Project Images**: `uploads/project/{projectId}`

### Linting and Code Style

Run these checks before committing your changes:

| Task                | Command               |
| ------------------- | --------------------- |
| **Check Style**     | `npm run style-check` |
| **Format Code**     | `npm run format`      |
| **Check Lint**      | `npm run lint`        |
| **Fix Lint Issues** | `npm run lint:fix`    |

### Recommended Extensions (VS Code)

- **Prettier** (`esbenp.prettier-vscode`)
- **Run on Save** (`pucelle.run-on-save`)
- **Proto Formatter** (`DrBlury.protobuf-vsc`)

---

## 🚢 DevOps & Deployment

### Production Environment Setup

To prepare for a production release, copy the example files and populate them with the target server credentials:

1. `db/.env.example` ➡️ `db/.env.prod`
2. `frontend/.env.example` ➡️ `frontend/.env.prod`
3. `backend/.env.example` ➡️ `backend/.env.prod`
