CREATE TABLE "user" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  password BYTEA NOT NULL,
  profile_picture TEXT NULL,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE "project" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE "user_to_project" (
  project_id UUID NOT NULL REFERENCES "project" (id),
  user_id UUID NOT NULL REFERENCES "user" (id),
  PRIMARY KEY (project_id, user_id)
);

CREATE TYPE camera AS (
  id UUID,
  angle_x DOUBLE PRECISION,
  angle_y DOUBLE PRECISION,
  angle_z DOUBLE PRECISION,
  pos_x DOUBLE PRECISION,
  pos_y DOUBLE PRECISION,
  pos_z DOUBLE PRECISION
);

CREATE TABLE "model" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES "project" (id) ON DELETE CASCADE,
  -- filename or model name, conflictable
  name VARCHAR(255) NOT NULL,
  -- description, mutable
  description TEXT NOT NULL DEFAULT '',
  -- storage location, mutable
  file_path TEXT NOT NULL,
  -- cameras info, mutable
  cameras camera[] DEFAULT '{}',
  -- version tracking
  version INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE "user_model_snapshots" (
  PRIMARY KEY (project_id, user_id),
  -- immutable
  user_id UUID NOT NULL REFERENCES "user" (id),
  -- immutable
  project_id UUID NOT NULL REFERENCES "project" (id) ON DELETE CASCADE,
  -- filename or model name, conflictable
  name VARCHAR(255) NOT NULL,
  -- description, mutable
  description TEXT NOT NULL DEFAULT '',
  -- storage location, mutable
  file_path TEXT NOT NULL,
  -- cameras info, mutable
  cameras camera[] DEFAULT '{}',
  -- version tracking
  version INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);
