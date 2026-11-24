package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// 가사집 검색 결과 화면의 구조만 옮긴 샘플 (가사는 저작권 때문에 지어낸 문장).
// 사이트 구조가 바뀌면 TestGasazipLive 로 확인하고 이 샘플도 같이 고친다.
const searchFixture = `<div class="list-group gz-candy-list gz-candy-search-results">
  <a href="/100" class="list-group-item gz-candy-row gz-candy-song-row">
    <span class="gz-candy-row-main">
      <strong>은혜의 강가</strong>
      <em>가수A</em>
      <span class="gz-candy-snippet"><em>은혜</em> 강가에 서서</span>
    </span>
  </a>
  <a href="/200" class="list-group-item gz-candy-row gz-candy-song-row">
    <span class="gz-candy-row-main">
      <strong>은혜의 강가 (Live)</strong>
      <em>가수B</em>
      <span class="gz-candy-snippet">강가에 서서</span>
    </span>
  </a>
  <a href="/300" class="list-group-item gz-candy-row gz-candy-song-row">
    <span class="gz-candy-row-main">
      <strong>다른 노래</strong>
      <em>가수C</em>
      <span class="gz-candy-snippet">첫 소절은 주 은혜 날 덮으니</span>
    </span>
  </a>
</div>`

const songFixture = `<div id="gasa-copy" aria-hidden="true">
첫째 줄<br />
둘째 줄<br />
<br />
셋째 줄<br />
</div>`

func mustDoc(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestFindSongPath(t *testing.T) {
	doc := mustDoc(t, searchFixture)
	cases := []struct{ name, title, artist, want string }{
		{"제목 앞부분 일치는 첫 곡", "은혜의 강가", "", "/100"},
		{"띄어쓰기 무시", "은혜의강가", "", "/100"},
		{"가수까지 적으면 그 가수", "은혜의 강가", "가수B", "/200"},
		{"가수가 없으면 제목 일치 곡", "은혜의 강가", "없는가수", "/100"},
		{"제목에 없으면 가사 미리보기", "주 은혜 날 덮으니", "", "/300"},
		{"아무 데도 없으면 빈 값", "전혀 없는 곡", "", ""},
		{"빈 검색어", "  ", "", ""},
	}
	for _, c := range cases {
		if got := findSongPath(doc, c.title, c.artist); got != c.want {
			t.Errorf("%s: findSongPath(%q, %q) = %q, want %q", c.name, c.title, c.artist, got, c.want)
		}
	}
}

func TestParseLyrics(t *testing.T) {
	si := &SlideData{}
	si.parseLyrics(mustDoc(t, songFixture))
	if want := "첫째 줄\n둘째 줄\n\n셋째 줄"; si.Lyrics != want {
		t.Errorf("Lyrics = %q, want %q", si.Lyrics, want)
	}
	if len(si.Content) == 0 {
		t.Error("Content 가 비어 있음")
	}

	empty := &SlideData{}
	empty.parseLyrics(mustDoc(t, `<div>no lyrics</div>`))
	if empty.Lyrics != "" {
		t.Errorf("가사 영역이 없으면 빈 값이어야 함: %q", empty.Lyrics)
	}
}

func TestSplitQuery(t *testing.T) {
	cases := []struct{ in, title, artist string }{
		{"은혜", "은혜", ""},
		{" 은혜 - 손경민 ", "은혜", "손경민"},
		{"A - B - C", "A - B", "C"},
		{"하나님-아버지", "하나님-아버지", ""},
	}
	for _, c := range cases {
		if title, artist := SplitQuery(c.in); title != c.title || artist != c.artist {
			t.Errorf("SplitQuery(%q) = (%q, %q), want (%q, %q)", c.in, title, artist, c.title, c.artist)
		}
	}
}

// TestGasazipLive — 실제 가사집에 접속해 검색·가사 읽기가 되는지 확인한다(사이트 구조 변경 감지).
// 네트워크가 필요해서 기본은 건너뛴다: EP_LIVE_LYRICS=1 go test ./internal/parser -run Live
func TestGasazipLive(t *testing.T) {
	if os.Getenv("EP_LIVE_LYRICS") != "1" {
		t.Skip("EP_LIVE_LYRICS=1 일 때만 실행")
	}
	si := &SlideData{}
	if err := si.SearchLyricsList("은혜 - 손경민"); err != nil {
		t.Fatalf("가사집 접속 실패: %v", err)
	}
	if si.Lyrics == "" {
		t.Fatal("가사를 못 읽음 — 검색 결과(a.gz-candy-song-row) 또는 가사 영역(#gasa-copy) 구조가 바뀌었는지 확인")
	}
}
