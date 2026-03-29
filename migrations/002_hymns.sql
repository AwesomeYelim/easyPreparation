-- 찬송가 테이블
CREATE TABLE IF NOT EXISTS hymns (
    id SERIAL PRIMARY KEY,
    hymnbook VARCHAR(20) NOT NULL DEFAULT 'new',
    number INT NOT NULL,
    title VARCHAR(200) NOT NULL,
    first_line VARCHAR(500),
    category VARCHAR(100),
    lyrics TEXT,
    has_pdf BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(hymnbook, number)
);
CREATE INDEX IF NOT EXISTS idx_hymns_title ON hymns(title);
CREATE INDEX IF NOT EXISTS idx_hymns_number ON hymns(number);
