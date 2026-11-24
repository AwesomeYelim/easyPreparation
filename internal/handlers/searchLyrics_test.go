package handlers

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// withLyricsDBs — 메모리 DB 로 custom_songs·hymns 를 만들어 apiDB·bibleDB 를 잠시 바꿔 끼운다.
func withLyricsDBs(t *testing.T) {
	t.Helper()
	open := func() *sql.DB {
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		db.SetMaxOpenConns(1) // :memory: 는 연결마다 다른 DB 라 하나로 고정
		return db
	}
	app, bible := open(), open()
	mustExec := func(db *sql.DB, q string, args ...interface{}) {
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatal(err)
		}
	}
	mustExec(app, `CREATE TABLE custom_songs (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL,
		artist TEXT DEFAULT '', lyrics TEXT NOT NULL, used_count INTEGER DEFAULT 0, last_used TEXT DEFAULT '')`)
	mustExec(app, `INSERT INTO custom_songs (title, artist, lyrics, used_count) VALUES
		('은혜 아니면', '', 'L-은혜아니면', 9),
		('은혜 - 손경민', '', 'L-은혜-손경민', 1),
		('주님의 마음', '어노인팅', 'L-주님의마음-어노인팅', 1)`)
	mustExec(bible, `CREATE TABLE hymns (hymnbook TEXT DEFAULT 'new', number INTEGER, title TEXT, first_line TEXT, lyrics TEXT)`)
	mustExec(bible, `INSERT INTO hymns (number, title, first_line, lyrics) VALUES
		(305, '나 같은 죄인 살리신', '', 'H-305'),
		(1, '만복의 근원 하나님', '', 'H-1')`)

	oldApp, oldBible := apiDB, bibleDB
	apiDB, bibleDB = app, bible
	t.Cleanup(func() {
		apiDB, bibleDB = oldApp, oldBible
		app.Close()
		bible.Close()
	})
}

func TestSearchCustomSongs(t *testing.T) {
	withLyricsDBs(t)
	cases := []struct{ query, want string }{
		{"은혜", "L-은혜-손경민"},          // 가수를 뺀 제목이 같은 곡 (일부만 맞는 '은혜 아니면' 은 안 씀)
		{"은혜 - 손경민", "L-은혜-손경민"},   // 제목·가수 모두 일치
		{"은혜 - 다른가수", ""},           // 가수가 다르면 안 씀
		{"은혜아니면", "L-은혜아니면"},       // 띄어쓰기 무시
		{"주님의 마음 - 어노인팅", "L-주님의마음-어노인팅"}, // artist 칸으로 등록된 곡
		{"주님의 마음", "L-주님의마음-어노인팅"},
		{"주님", ""}, // 일부만 맞으면 안 씀
	}
	for _, c := range cases {
		if got := searchCustomSongs(c.query); got != c.want {
			t.Errorf("searchCustomSongs(%q) = %q, want %q", c.query, got, c.want)
		}
	}
}

func TestSearchHymnDB(t *testing.T) {
	withLyricsDBs(t)
	cases := []struct{ query, want string }{
		{"305", "H-305"},
		{"305장", "H-305"},
		{"새찬송가 305장 나 같은 죄인", "H-305"},
		{"찬송가 1장", "H-1"},
		{"나 같은 죄인 살리신", "H-305"},
		{"나같은죄인살리신", "H-305"},
		{"나 같은 죄인", ""}, // 제목 일부는 안 씀 (가사집으로 넘어감)
		{"하나님", ""},
		{"10000 Reasons", ""}, // 숫자로 시작하는 CCM 제목은 장 번호가 아님
	}
	for _, c := range cases {
		if got := searchHymnDB(c.query); got != c.want {
			t.Errorf("searchHymnDB(%q) = %q, want %q", c.query, got, c.want)
		}
	}
}
