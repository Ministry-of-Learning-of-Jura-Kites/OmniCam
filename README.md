# Setup

## Core Tools Version

Nodejs -> v22.18.0

Golang -> v1.24.5

## CLI

### sqlc

To generate golang type interfaces and queries handler from sql

`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

### golang-migrate

To migrate database to newer version

`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

### swagger-ui-cli

To show swagger to test

`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

## Env

Create a file frontend/.env and backend/.env, you can see meaning of each env in .env.example

\*\* When adding env, make sure to add to schema and not use os.GetEnv or process.env

## Extension

### Prettier (esbenp.prettier-vscode)

### Run on Save (pucelle.run-on-save)

# Running

Frontend Dev -> `npm run dev`

Swagger -> `npm run swagger`

Backend -> `go ./backend/main.go`

# Database

We use PostgreSQL as a database, to spin up postgres open a new terminal, and use command

`docker run --rm -it --name omnicam-postgres -e POSTGRES_PASSWORD=password -p 5433:5432 postgres:17`

## Schema

When adding a schema, make sure to add both .up.sql and .down.sql, up is for applying migration, and down is for rollbacking migration

\*\*\* Order of commands must be reversed for down!

## Query

Follow query format of sqlc

\* When saving a schema or query file sqlc generate should be run automatically with extension Run on Save

# Lint and Style

Please run style check and lint before commit if possible

To check style

`npm run style-check`

To format

`npm run format`

To check lint

`npm run lint`

To fix lint

`npm run lint:fix`
