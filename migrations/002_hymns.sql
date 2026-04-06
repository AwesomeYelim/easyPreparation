-- 찬송가 테이블 (마스터 스키마: data/schema.sql)
CREATE TABLE IF NOT EXISTS hymns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hymnbook TEXT NOT NULL DEFAULT 'new',
    number INTEGER NOT NULL,
    title TEXT NOT NULL,
    first_line TEXT DEFAULT '',
    category TEXT DEFAULT '',
    lyrics TEXT DEFAULT '',
    has_pdf INTEGER DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now')),
    UNIQUE(hymnbook, number)
);
CREATE INDEX IF NOT EXISTS idx_hymns_title ON hymns(title);
CREATE INDEX IF NOT EXISTS idx_hymns_number ON hymns(number);
