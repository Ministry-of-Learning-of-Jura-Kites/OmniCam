CREATE TABLE "user" (
  id UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
  email VARCHAR(255) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  password BYTEA NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE "project" (
  id UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
  name VARCHAR(255) NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  image_path TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE "user_to_project" (
  project_id UUID NOT NULL REFERENCES "project" (id),
  user_id UUID NOT NULL REFERENCES "user" (id),
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
  -- version tracking
  version INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- snapshots of models per user
CREATE TABLE "user_model_workspace" (
  PRIMARY KEY (model_id, user_id),
  user_id UUID NOT NULL REFERENCES "user" (id),
  -- reference back to model
  model_id UUID NOT NULL REFERENCES "model" (id) ON DELETE CASCADE,
  -- filename or model name
  name VARCHAR(255) NOT NULL,
  -- description, mutable
  description TEXT NOT NULL DEFAULT '',
  -- store cameras as a document
  cameras JSONB NOT NULL DEFAULT '{}'::JSONB,
  -- storage location, mutable
  file_path TEXT NOT NULL,
  -- version tracking
  version INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- CREATE TABLE "camera" (
--   id UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
--   name VARCHAR(255) NOT NULL,
--   angle_x DOUBLE PRECISION,
--   angle_y DOUBLE PRECISION,
--   angle_z DOUBLE PRECISION,
--   angle_w DOUBLE PRECISION,
--   pos_x DOUBLE PRECISION,
--   pos_y DOUBLE PRECISION,
--   pos_z DOUBLE PRECISION,
--   -- always belongs to a model
--   model_id UUID NOT NULL REFERENCES "model" (id) ON DELETE CASCADE,
--   -- optionally belongs to a user (snapshot case)
--   user_id UUID REFERENCES "user" (id) ON DELETE CASCADE,
--   -- auto derived: true if snapshot, false if model camera
--   is_snapshot BOOLEAN generated always AS (user_id IS NOT NULL) stored,
--   -- if user_id is not null → must exist in user_model_snapshots
--   CONSTRAINT fk_camera_snapshot FOREIGN key (model_id, user_id) REFERENCES "user_model_snapshots" (model_id, user_id) ON DELETE CASCADE
-- );
