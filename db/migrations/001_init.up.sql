CREATE TABLE "user" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  name VARCHAR(255),
  created_at TIMESTAMPTZ DEFAULT now (),
  updated_at TIMESTAMPTZ DEFAULT now ()
);

CREATE TABLE "project" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  project_name VARCHAR(255) NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ DEFAULT now (),
  updated_at TIMESTAMPTZ DEFAULT now ()
);

CREATE TABLE "model" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  project_id UUID REFERENCES "Project" (id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL, -- filename or model name
  file_path TEXT NOT NULL, -- storage location
  camera_angle_x FLOAT, -- rotation angles (optional)
  camera_angle_y FLOAT,
  camera_angle_z FLOAT,
  camera_pos_x FLOAT, -- camera position in 3D space
  camera_pos_y FLOAT,
  camera_pos_z FLOAT,
  version_id UUID DEFAULT gen_random_uuid (), -- version tracking
  created_at TIMESTAMPTZ DEFAULT now (),
  updated_at TIMESTAMPTZ DEFAULT now ()
);
