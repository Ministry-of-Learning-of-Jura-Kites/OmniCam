INSERT INTO
  "user" (
    id,
    email,
    username,
    first_name,
    last_name,
    password
  )
VALUES
  (
    '00000000-0000-0000-0000-000000000000',
    'null@example.com',
    'username',
    'first name',
    'last name',
    ''::BYTEA
  );

INSERT INTO
  "project" (id, name, description, image_path)
VALUES
  (
    '00000000-0000-0000-0000-000000000000',
    'null_project',
    '',
    ''
  );

INSERT INTO
  "user_to_project" (project_id, role, user_id)
VALUES
  (
    '00000000-0000-0000-0000-000000000000',
    'owner',
    '00000000-0000-0000-0000-000000000000'
  );

INSERT INTO
  "model" (
    id,
    project_id,
    name,
    description,
    cameras,
    file_path,
    image_path,
    version
  )
VALUES
  (
    '00000000-0000-0000-0000-000000000000',
    '00000000-0000-0000-0000-000000000000',
    'null_model',
    '',
    '{}'::JSONB,
    '',
    '',
    0
  );

INSERT INTO
  user_model_workspace (
    name,
    user_id,
    model_id,
    file_path,
    description,
    cameras,
    base_cameras,
    version,
    base_version,
    created_at,
    updated_at
  )
VALUES
  (
    'g',
    '00000000-0000-0000-0000-000000000000',
    '00000000-0000-0000-0000-000000000000',
    'g',
    'g',
    '{}'::JSONB,
    '{}'::JSONB,
    0,
    0,
    NOW(),
    NOW()
  );
