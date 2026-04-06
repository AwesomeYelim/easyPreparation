-- =============================================================================
-- easyPreparation SQLite Schema
-- PostgreSQL 마이그레이션(001~004)을 SQLite 호환 DDL로 변환
-- =============================================================================
-- PRAGMA 설정
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

-- =============================================================================
-- 1. bible_versions
--    성경 번역본 목록 (개역개정 id=1 기준)
-- =============================================================================
CREATE TABLE IF NOT EXISTS bible_versions (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(50)  NOT NULL,
    code VARCHAR(10)  NOT NULL UNIQUE
);

-- 기본 번역본 데이터 삽입 (migration 001)
INSERT INTO bible_versions (id, name, code) VALUES
    (1, '개역개정', 'KRV2'),
    (2, '개역한글', 'KRV'),
    (3, '공동번역', 'COMMON'),
    (4, '표준새번역', 'NKSV'),
    (5, 'NIV', 'NIV'),
    (6, 'KJV', 'KJV'),
    (7, '우리말성경', 'URNB')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- 2. books
--    성경 책 목록 (66권)
-- =============================================================================
CREATE TABLE IF NOT EXISTS books (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name_kor   VARCHAR(20)  NOT NULL,
    name_eng   VARCHAR(50)  NOT NULL,
    abbr_kor   VARCHAR(10)  NOT NULL,
    abbr_eng   VARCHAR(10)  NOT NULL,
    book_order INTEGER      NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_books_order ON books(book_order);

-- =============================================================================
-- 3. verses
--    성경 본문 (버전 × 책 × 장 × 절 복합 PK)
-- =============================================================================
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

-- =============================================================================
-- 4. churches
--    계정(교회) 테이블 — apiHandlers.go의 INSERT/SELECT 기준
-- =============================================================================
CREATE TABLE IF NOT EXISTS churches (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    name         VARCHAR(200) NOT NULL DEFAULT '',
    english_name VARCHAR(200) NOT NULL DEFAULT '',
    email        VARCHAR(255) NOT NULL UNIQUE,
    created_at   DATETIME     NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_churches_email ON churches(email);

-- =============================================================================
-- 5. licenses
--    라이선스(Figma 키/토큰 포함) — settingsHandlers.go LicenseHandler 기준
-- =============================================================================
CREATE TABLE IF NOT EXISTS licenses (
    church_id     INTEGER  NOT NULL UNIQUE REFERENCES churches(id) ON DELETE CASCADE,
    license_key   VARCHAR(500),
    license_token VARCHAR(500),
    plan          VARCHAR(50),
    expires_at    DATETIME,
    device_id     VARCHAR(200),
    issued_at     DATETIME NOT NULL DEFAULT (datetime('now')),
    last_verified TEXT,
    signature     TEXT     DEFAULT ''
);

-- =============================================================================
-- 6. hymns
--    찬송가 테이블 (migration 002) — SERIAL → AUTOINCREMENT, BOOLEAN → INTEGER
-- =============================================================================
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

-- =============================================================================
-- 7. user_settings
--    사용자(교회별) 설정 (migration 003) — JSONB → TEXT
-- =============================================================================
CREATE TABLE IF NOT EXISTS user_settings (
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    church_id               INTEGER NOT NULL UNIQUE REFERENCES churches(id) ON DELETE CASCADE,
    preferred_bible_version INTEGER NOT NULL DEFAULT 1,
    theme                   VARCHAR(10)  NOT NULL DEFAULT 'light',
    font_size               INTEGER NOT NULL DEFAULT 16,
    default_bpm             INTEGER NOT NULL DEFAULT 100,
    display_layout          VARCHAR(20)  NOT NULL DEFAULT 'default',
    custom_worship_template TEXT,
    updated_at              DATETIME NOT NULL DEFAULT (datetime('now'))
);

-- =============================================================================
-- 8. generation_history
--    생성 이력 (migration 003 + 004 — order_data 컬럼 포함)
--    JSONB → TEXT, SERIAL → AUTOINCREMENT
-- =============================================================================
CREATE TABLE IF NOT EXISTS generation_history (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    church_id  INTEGER      NOT NULL REFERENCES churches(id) ON DELETE CASCADE,
    type       VARCHAR(30)  NOT NULL,
    filename   VARCHAR(500),
    status     VARCHAR(20)  NOT NULL DEFAULT 'success',
    metadata   TEXT,
    file_path  VARCHAR(500),
    order_data TEXT,
    created_at DATETIME     NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_gen_history_church
    ON generation_history(church_id, created_at DESC);
