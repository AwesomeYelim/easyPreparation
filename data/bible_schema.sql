-- =============================================================================
-- Bible & Hymn Database Schema (bible.db)
-- 성경 + 찬송가 전용 DB — 앱과 독립적으로 사용 가능
-- =============================================================================
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

-- 1. bible_versions — 성경 번역본 목록
CREATE TABLE IF NOT EXISTS bible_versions (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(50)  NOT NULL,
    code VARCHAR(10)  NOT NULL UNIQUE
);

INSERT INTO bible_versions (id, name, code) VALUES
    (1, '개역개정', 'KRV2'),
    (2, '개역한글', 'KRV'),
    (3, '공동번역', 'COMMON'),
    (4, '표준새번역', 'NKSV'),
    (5, 'NIV', 'NIV'),
    (6, 'KJV', 'KJV'),
    (7, '우리말성경', 'URNB')
ON CONFLICT (id) DO NOTHING;

-- 2. books — 성경 책 목록 (66권)
CREATE TABLE IF NOT EXISTS books (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name_kor   VARCHAR(20)  NOT NULL,
    name_eng   VARCHAR(50)  NOT NULL,
    abbr_kor   VARCHAR(10)  NOT NULL,
    abbr_eng   VARCHAR(10)  NOT NULL,
    book_order INTEGER      NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_books_order ON books(book_order);

-- 3. verses — 성경 본문
CREATE TABLE IF NOT EXISTS verses (
    version_id INTEGER NOT NULL REFERENCES bible_versions(id) ON DELETE CASCADE,
    book_id    INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    chapter    INTEGER NOT NULL,
    verse      INTEGER NOT NULL,
    text       TEXT    NOT NULL,
    PRIMARY KEY (version_id, book_id, chapter, verse)
);

CREATE INDEX IF NOT EXISTS idx_verses_lookup
    ON verses(version_id, book_id, chapter, verse);

-- 4. hymns — 찬송가
CREATE TABLE IF NOT EXISTS hymns (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    hymnbook   VARCHAR(20)  NOT NULL DEFAULT 'new',
    number     INTEGER      NOT NULL,
    title      VARCHAR(200) NOT NULL,
    first_line VARCHAR(500),
    category   VARCHAR(100),
    lyrics     TEXT,
    has_pdf    INTEGER      NOT NULL DEFAULT 0,
    created_at DATETIME     NOT NULL DEFAULT (datetime('now')),
    UNIQUE(hymnbook, number)
);

CREATE INDEX IF NOT EXISTS idx_hymns_title  ON hymns(title);
CREATE INDEX IF NOT EXISTS idx_hymns_number ON hymns(number);
