ALTER TABLE user_model_workspace
ADD COLUMN simulation JSONB NOT NULL DEFAULT '{}'::JSONB;
