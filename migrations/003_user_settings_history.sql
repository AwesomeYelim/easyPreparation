-- 사용자 설정 테이블
CREATE TABLE IF NOT EXISTS user_settings (
    id SERIAL PRIMARY KEY,
    church_id INT NOT NULL REFERENCES churches(id) ON DELETE CASCADE,
    preferred_bible_version INT DEFAULT 1,
    theme VARCHAR(10) DEFAULT 'light',
    font_size INT DEFAULT 16,
    default_bpm INT DEFAULT 100,
    display_layout VARCHAR(20) DEFAULT 'default',
    custom_worship_template JSONB,
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(church_id)
);

-- 생성 이력 테이블
CREATE TABLE IF NOT EXISTS generation_history (
    id SERIAL PRIMARY KEY,
    church_id INT NOT NULL REFERENCES churches(id) ON DELETE CASCADE,
    type VARCHAR(30) NOT NULL,
    filename VARCHAR(500),
    status VARCHAR(20) DEFAULT 'success',
    metadata JSONB,
    file_path VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_gen_history_church ON generation_history(church_id, created_at DESC);
