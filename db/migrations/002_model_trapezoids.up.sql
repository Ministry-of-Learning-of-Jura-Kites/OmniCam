ALTER TABLE "model"
ADD COLUMN target_area_trapezoids JSONB NOT NULL DEFAULT '{}'::JSONB;

ALTER TABLE "user_model_workspace"
ADD COLUMN target_area_trapezoids JSONB NOT NULL DEFAULT '{}'::JSONB;
