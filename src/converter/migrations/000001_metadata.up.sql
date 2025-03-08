CREATE TABLE IF NOT EXISTS metadata (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    video_key VARCHAR(255) NOT NULL,
    audio_key VARCHAR(255) NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
