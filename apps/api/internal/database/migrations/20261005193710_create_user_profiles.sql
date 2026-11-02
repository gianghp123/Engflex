-- +goose Up
CREATE TABLE user_profiles (
    id                    uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    clerk_user_id         text PRIMARY KEY,
    level text NOT NULL DEFAULT 'beginner' CHECK (level IN ('beginner','intermediate','advanced')),
    goals text[] NOT NULL DEFAULT '{}' CHECK (array_position(goals,NULL) IS NULL AND goals <@ ARRAY['travel','work','exams','daily','fun']::text[]),
    topics text[] NOT NULL DEFAULT '{}' CHECK (array_position(topics,NULL) IS NULL AND topics <@ ARRAY['travel','business','movies_tv','technology','food','music','sports','news']::text[]),
    daily_commitment_min integer NOT NULL DEFAULT 15 CHECK (daily_commitment_min IN (5,10,15,30)),
    onboarding_completed_at timestamptz,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS user_profiles;
