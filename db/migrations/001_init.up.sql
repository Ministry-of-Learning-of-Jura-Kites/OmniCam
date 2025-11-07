CREATE TYPE role AS ENUM('owner', 'project_manager', 'collaborator');

CREATE TABLE "user" (
  id UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
  email VARCHAR(255) NOT NULL UNIQUE,
  first_name VARCHAR(255) NOT NULL,
  username VARCHAR(255) NOT NULL UNIQUE,
  last_name VARCHAR(255) NOT NULL,
  password BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "project" (
  id UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
  name VARCHAR(255) NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  image_path TEXT NOT NULL,
  file_extension TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "user_to_project" (
  project_id UUID NOT NULL REFERENCES "project" (id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
  role role NOT NULL,
  PRIMARY KEY (project_id, user_id)
);

CREATE TABLE "model" (
  id UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
  project_id UUID NOT NULL REFERENCES "project" (id) ON DELETE CASCADE,
  -- filename or model name, conflictable
  name VARCHAR(255) NOT NULL,
  -- description, mutable
  description TEXT NOT NULL DEFAULT '',
  -- store cameras as a document
  cameras JSONB NOT NULL DEFAULT '{}'::JSONB,
  -- storage location, mutable
  file_path TEXT NOT NULL,
  image_path TEXT NOT NULL,
  model_extension TEXT NOT NULL,
  image_extension TEXT NOT NULL,
  -- version tracking
  version INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- snapshots of models per user
CREATE TABLE "user_model_workspace" (
  PRIMARY KEY (model_id, user_id),
  user_id UUID NOT NULL REFERENCES "user" (id),
  -- reference back to model
  model_id UUID NOT NULL REFERENCES "model" (id) ON DELETE CASCADE,
  -- store cameras as a document
  cameras JSONB NOT NULL DEFAULT '{}'::JSONB,
  -- store branched-out cameras as a document
  base_cameras JSONB NOT NULL DEFAULT '{}'::JSONB,
  -- version tracking
  version INT NOT NULL DEFAULT 0,
  -- branched-out version
  base_version INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
