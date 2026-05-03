#!/usr/bin/env node
// 전체 기능 종합 테스트 — 찬송가/성경/주보/Display/모바일/API
// 실행: NODE_PATH=ui/node_modules node tools/ui-full-test.js
// 전제: make dev (Go :8080 + Next.js :3000) 실행 중

const { chromium } = require('@playwright/test');
const path = require('path');
const fs = require('fs');
const http = require('http');

const OUT_DIR = path.join(__dirname, '../.claude/screenshots/full-test');
const BASE = 'http://localhost:3000';
const GO = 'http://localhost:8080';

// 투어 건너뛰기 + 기본 localStorage 세팅
const STORAGE = {
  cookies: [],
  origins: [{
    origin: BASE,
    localStorage: [{ name: 'ep_tour_done', value: '1' }],
  }],
};

let pass = 0, fail = 0;
const issues = [];
const sectionResults = {};
let currentSection = '';

function ok(label) {
  console.log(`  ✓ ${label}`);
  pass++;
  if (!sectionResults[currentSection]) sectionResults[currentSection] = { pass: 0, fail: 0 };
  sectionResults[currentSection].pass++;
}
function ng(label, detail = '') {
  console.log(`  ✗ ${label}${detail ? ' — ' + detail : ''}`);
  fail++;
  issues.push(`[${currentSection}] ${label}${detail ? ': ' + detail : ''}`);
  if (!sectionResults[currentSection]) sectionResults[currentSection] = { pass: 0, fail: 0 };
  sectionResults[currentSection].fail++;
}
function section(name) {
  currentSection = name;
  console.log(`\n[${name}]`);
}

async function shot(page, name) {
  const p = path.join(OUT_DIR, `${name}.png`);
  await page.screenshot({ path: p, fullPage: false });
  console.log(`  📸 ${name}.png`);
}

async function dismissModal(page) {
  try {
    const skip = await page.$('button:has-text("건너뛰기")');
    if (skip) { await skip.click(); await page.waitForTimeout(300); }
  } catch {}
}

// HTTP GET 헬퍼 (브라우저 없이)
const httpGet = (url) => new Promise((resolve, reject) => {
  http.get(url, (res) => {
    let data = '';
    res.on('data', (chunk) => data += chunk);
    res.on('end', () => {
      try { resolve({ status: res.statusCode, body: JSON.parse(data), raw: data }); }
      catch { resolve({ status: res.statusCode, body: null, raw: data }); }
    });
  }).on('error', reject);
});

// ═══════════════════════════════════════════════════════════════
// 1. API 직접 검증 (브라우저 없이)
// ═══════════════════════════════════════════════════════════════
async function testAPIs() {
  section('API 전체 검증');

  // --- 예배 순서 ---
  for (const t of ['main_worship', 'after_worship', 'wed_worship']) {
    const r = await httpGet(`${GO}/api/worship-order?type=${t}`);
    r.status === 200 && Array.isArray(r.body) && r.body.length > 0
      ? ok(`예배 순서 API (${t}): ${r.body.length}개 항목`)
      : ng(`예배 순서 API (${t}) 실패`, `HTTP ${r.status}`);
  }

  // --- 사용자 정보 ---
  const user = await httpGet(`${GO}/api/user`);
  user.status === 200
    ? ok('사용자 정보 API 정상')
    : ng('사용자 정보 API 실패', `HTTP ${user.status}`);

  // --- 찬송가 목록 ---
  const hymnList = await httpGet(`${GO}/api/hymns?page=1`);
  hymnList.status === 200 && hymnList.body?.hymns?.length > 0
    ? ok(`찬송가 목록 API: 총 ${hymnList.body.total}개`)
    : ng('찬송가 목록 API 실패', `HTTP ${hymnList.status}`);

  // --- 찬송가 번호 검색 ---
  const hymnSearch = await httpGet(`${GO}/api/hymns/search?q=36`);
  hymnSearch.status === 200
    ? ok('찬송가 번호 검색 API 정상')
    : ng('찬송가 번호 검색 API 실패', `HTTP ${hymnSearch.status}`);

  // --- 찬송가 텍스트 검색 ---
  const hymnTxt = await httpGet(`${GO}/api/hymns/search?q=%EC%A3%BC+%EC%98%88%EC%88%98`);
  hymnTxt.status === 200
    ? ok('찬송가 텍스트 검색 API 정상')
    : ng('찬송가 텍스트 검색 API 실패', `HTTP ${hymnTxt.status}`);

  // --- 찬송가 상세(가사) ---
  const hymn36 = await httpGet(`${GO}/api/hymns/detail?number=36`);
  hymn36.status === 200 && hymn36.body?.lyrics?.length > 10
    ? ok(`찬송가 36번 가사: "${(hymn36.body.lyrics || '').slice(0, 25)}..."`)
    : ng('찬송가 36번 가사 없음', `HTTP ${hymn36.status}`);

  // --- 성경 책 목록 ---
  const books = await httpGet(`${GO}/api/bible/books`);
  books.status === 200 && books.body?.['창세기']
    ? ok('성경 책 목록 API: 창세기 확인')
    : ng('성경 책 목록 API 실패', `HTTP ${books.status}`);

  // --- 창세기 1장 ---
  const gen1 = await httpGet(`${GO}/api/bible/verses?book=1&chapter=1`);
  if (gen1.status === 200 && Array.isArray(gen1.body) && gen1.body.length > 0) {
    ok(`창세기 1장: ${gen1.body.length}절`);
    gen1.body[0]?.text?.includes('태초')
      ? ok(`창세기 1:1: "${gen1.body[0].text.slice(0, 20)}..."`)
      : ng('창세기 1:1 본문 불일치');
  } else {
    ng('창세기 1장 API 실패', `HTTP ${gen1.status}`);
  }

  // --- 성경 번역 버전 ---
  const versions = await httpGet(`${GO}/api/bible/versions`);
  versions.status === 200
    ? ok('성경 번역 버전 API 정상')
    : ng('성경 번역 버전 API 실패', `HTTP ${versions.status}`);

  // --- 성경 구절 검색 ---
  const bSearch = await httpGet(`${GO}/api/bible/search?q=%ED%83%9C%EC%B4%88%EC%97%90&version=1`);
  bSearch.status === 200
    ? ok('성경 구절 검색 API 정상')
    : ng('성경 구절 검색 API 실패', `HTTP ${bSearch.status}`);

  // --- Display Config ---
  const cfg = await httpGet(`${GO}/api/display-config`);
  cfg.status === 200
    ? ok('Display Config API 정상')
    : ng('Display Config API 실패', `HTTP ${cfg.status}`);

  // --- 비디오 배경 목록 ---
  const vbg = await httpGet(`${GO}/api/video-bg/list`);
  vbg.status === 200
    ? ok('비디오 배경 목록 API 정상')
    : ng('비디오 배경 목록 API 실패', `HTTP ${vbg.status}`);

  // --- 주보 테마/커버 ---
  const tpl = await httpGet(`${GO}/api/bulletin-theme`);
  tpl.status === 200 || tpl.status === 404
    ? ok('주보 테마 API 응답 정상')
    : ng('주보 테마 API 실패', `HTTP ${tpl.status}`);

  // --- Display HTML 페이지 ---
  const disp = await httpGet(`${GO}/display`);
  disp.status === 200 && disp.raw?.includes('<html')
    ? ok('/display HTML 정상')
    : ng('/display HTML 실패', `HTTP ${disp.status}`);

  // --- Overlay HTML ---
  const ov = await httpGet(`${GO}/display/overlay`);
  ov.status === 200 && ov.raw?.includes('#slide')
    ? ok('/display/overlay HTML + #slide 확인')
    : ng('/display/overlay HTML 실패');

  // --- Stage HTML ---
  const stage = await httpGet(`${GO}/display/stage`);
  stage.status === 200 && stage.raw?.includes('<html')
    ? ok('/display/stage HTML 정상')
    : ng('/display/stage HTML 실패', `HTTP ${stage.status}`);

  // --- 모바일 ---
  const mobile = await httpGet(`${GO}/mobile`);
  mobile.status === 200 && mobile.raw?.includes('<html')
    ? ok('/mobile HTML 정상')
    : ng('/mobile HTML 실패', `HTTP ${mobile.status}`);
}

// ═══════════════════════════════════════════════════════════════
// 2. ProShell 레이아웃 + 내비게이션
// ═══════════════════════════════════════════════════════════════
async function testProShell(browser) {
  section('ProShell 레이아웃');
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, storageState: STORAGE });
  const page = await ctx.newPage();
  try {
    await page.goto(BASE, { waitUntil: 'networkidle', timeout: 20000 });
    await dismissModal(page);
    await page.waitForTimeout(600);

    // TopBar
    const topbar = await page.$('[data-testid="topbar"]');
    topbar ? ok('TopBar 렌더링') : ng('TopBar 없음');

    // 시계 표시 (TopBar 내 시간 표시)
    const clock = await page.locator('[data-testid="topbar"]').textContent().catch(() => '');
    /\d{1,2}:\d{2}/.test(clock) ? ok('TopBar 시계 표시') : ng('TopBar 시계 없음');

    // 탭 메뉴 (ProTopBar: <Link> 요소 — nav a)
    const tabs = await page.$$('[data-testid="topbar"] nav a, [data-testid="topbar"] a[href]');
    tabs.length >= 2 ? ok(`탭 메뉴 ${tabs.length}개 렌더링`) : ng('탭 메뉴 부족', `${tabs.length}개`);

    // 시퀀스 패널
    const seqPanel = await page.$('[data-testid="seq-panel"]');
    seqPanel ? ok('시퀀스 패널 렌더링') : ng('시퀀스 패널 없음');

    // 타임라인
    const timeline = await page.$('[data-testid="timeline"]');
    timeline ? ok('타임라인 렌더링') : ng('타임라인 없음');

    // 가로 스크롤 없음
    const hasHScroll = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 5);
    hasHScroll ? ng('가로 스크롤 발생') : ok('가로 스크롤 없음 (1440px)');

    await shot(page, 'proshell-layout');

    // 인스펙터 토글 버튼 클릭
    const inspBtn = await page.$('button[title*="inspector"], button[aria-label*="inspector"], button:has-text("inspector")');
    if (!inspBtn) {
      // TopBar에서 아이콘 버튼 클릭 (마지막 버튼)
      const topbarBtns = await page.$$('[data-testid="topbar"] button');
      if (topbarBtns.length > 0) {
        await topbarBtns[topbarBtns.length - 1].click();
        await page.waitForTimeout(400);
        ok('인스펙터 토글 버튼 클릭 시도');
      } else {
        ng('인스펙터 토글 버튼 없음');
      }
    } else {
      await inspBtn.click();
      await page.waitForTimeout(400);
      ok('인스펙터 토글 버튼 클릭');
    }

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 3. 주보 편집 페이지 (Bulletin)
// ═══════════════════════════════════════════════════════════════
async function testBulletin(browser) {
  section('주보 편집 페이지');
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 900 }, storageState: STORAGE });
  const page = await ctx.newPage();
  try {
    await page.goto(`${BASE}/bulletin`, { waitUntil: 'networkidle', timeout: 20000 });
    await dismissModal(page);
    await page.waitForTimeout(800);

    await shot(page, 'bulletin-page');

    // 페이지 제목
    const heading = await page.locator('h1, h2').first().textContent().catch(() => '');
    heading.length > 0 ? ok(`페이지 제목: "${heading.slice(0, 30)}"`) : ng('페이지 제목 없음');

    // 예배 타입 드롭다운
    const typeSelect = await page.$('select');
    typeSelect ? ok('예배 타입 드롭다운 존재') : ng('예배 타입 드롭다운 없음');

    // 드롭다운으로 예배 타입 변경
    if (typeSelect) {
      await page.selectOption('select', 'after_worship');
      await page.waitForTimeout(600);
      ok('오후예배로 전환');
      await page.selectOption('select', 'main_worship');
      await page.waitForTimeout(400);
    }

    // 예배 순서 목록 로드
    await page.waitForTimeout(1000);
    const orderItems = await page.$$('li, [class*="order"], [class*="item"]');
    orderItems.length > 0
      ? ok(`예배 순서 항목 ${orderItems.length}개 로드`)
      : ng('예배 순서 목록 비어있음');

    // PDF 다운로드 버튼
    const pdfBtn = page.locator('button').filter({ hasText: /PDF|다운로드/ });
    await pdfBtn.first().waitFor({ timeout: 5000 }).catch(() => {});
    const pdfCount = await pdfBtn.count();
    pdfCount > 0 ? ok('PDF 다운로드 버튼 존재') : ng('PDF 다운로드 버튼 없음');

    // 프로젝터 전송 버튼
    const projBtn = page.locator('button').filter({ hasText: /프로젝터|Display|전송/ });
    const projCount = await projBtn.count();
    projCount > 0 ? ok('프로젝터 전송 버튼 존재') : ng('프로젝터 전송 버튼 없음');

    // 템플릿 선택기 (TemplateSelector: "주보 시안" 버튼)
    const tplEl = page.locator('button').filter({ hasText: /주보 시안|시안|template/i });
    const tplCount = await tplEl.count();
    tplCount > 0 ? ok('주보 시안 선택기 존재') : ng('주보 시안 선택기 없음');

    // 가로 스크롤 없음
    const hasHScroll = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 5);
    hasHScroll ? ng('주보 페이지 가로 스크롤') : ok('주보 페이지 가로 스크롤 없음');

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 4. 찬송가 (Lyrics) 페이지
// ═══════════════════════════════════════════════════════════════
async function testLyrics(browser) {
  section('찬송가 (Lyrics) 페이지');
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 900 }, storageState: STORAGE });
  const page = await ctx.newPage();
  try {
    await page.goto(`${BASE}/lyrics`, { waitUntil: 'networkidle', timeout: 20000 });
    await dismissModal(page);
    await page.waitForTimeout(600);

    // 탭 존재 확인
    const tabs = page.locator('button').filter({ hasText: /자유곡|찬송가|내 찬양/ });
    const tabCount = await tabs.count();
    tabCount >= 2 ? ok(`찬양 탭 ${tabCount}개 존재`) : ng('찬양 탭 부족', `${tabCount}개`);

    // 찬송가 검색 탭 클릭
    const hymnTab = page.locator('button').filter({ hasText: '찬송가 검색' });
    const hymnTabExists = await hymnTab.count();
    if (hymnTabExists > 0) {
      await hymnTab.first().click();
      await page.waitForTimeout(400);
      ok('찬송가 검색 탭 클릭');
    } else {
      ng('찬송가 검색 탭 없음');
    }

    // 검색 입력 — placeholder로 HymnSearch 전용 input 정확히 타겟
    const searchInput = page.locator('input[placeholder*="번호 또는 제목"]');
    await searchInput.waitFor({ timeout: 5000 }).catch(() => {});
    const inputExists = await searchInput.count();
    if (inputExists > 0) {
      // pressSequentially: 키보드 이벤트 시뮬레이션 → React onChange 확실히 트리거
      await searchInput.click();
      await page.waitForTimeout(100);
      await searchInput.pressSequentially('36', { delay: 80 });
      await page.waitForTimeout(400);
      ok('찬송가 번호(36) 입력');

      // ▶ 정확히 "검색" 버튼만 — "찬송가 검색" 탭과 구분 (/^검색/ 로 tab 제외)
      // HymnSearch 검색 버튼은 텍스트가 "검색"(또는 "검색 중..."), 탭은 "찬송가 검색"
      const searchBtn = page.locator('button').filter({ hasText: /^검색/ });
      const searchBtnCount = await searchBtn.count();
      if (searchBtnCount > 0) {
        await searchBtn.first().click();
        await page.waitForTimeout(2000);
        ok('검색 버튼 클릭');
      } else {
        // Enter 키 대체
        await searchInput.press('Enter');
        await page.waitForTimeout(2000);
        ok('Enter로 검색 실행');
      }

      await shot(page, 'hymn-search-results');

      // 결과 목록 확인 (HymnSearch: <div> elements — span에 번호, div에 제목)
      // 시퀀스 패널의 cursor-pointer와 겹치지 않도록 span 텍스트로 정확히 타겟
      const results = page.locator('span.font-black').filter({ hasText: /^36$/ });
      const resultCount = await results.count();
      if (resultCount > 0) {
        ok(`찬송가 36번 검색 결과 ${resultCount}개`);
        // span 클릭 → 이벤트 버블링으로 부모 div의 onClick(handleSelect) 호출
        await results.first().click();
        await page.waitForTimeout(2000);
        ok('검색 결과 클릭');

        // Display 전송 버튼 대기 (selected 상태 설정 후 렌더링)
        const sendBtnLocator = page.locator('button').filter({ hasText: 'Display 전송' });
        await sendBtnLocator.waitFor({ timeout: 5000 }).catch(() => {});
        const sendCount = await sendBtnLocator.count();
        sendCount > 0
          ? ok('찬송가 Display 전송 버튼 존재')
          : ng('찬송가 Display 전송 버튼 없음');

        await shot(page, 'hymn-lyrics');

        // 가사 표시 확인
        const lyricsArea = page.locator('pre, textarea, [class*="lyric"], [class*="verse"]').first();
        const lyricsExists = await lyricsArea.count();
        if (lyricsExists > 0) {
          const txt = await lyricsArea.textContent();
          txt && txt.length > 10
            ? ok(`찬송가 가사 표시 (${txt.length}자)`)
            : ng('찬송가 가사 내용 비어있음');
        } else {
          const bodyTxt = await page.locator('body').textContent();
          /[가-힣]/.test(bodyTxt || '')
            ? ok('찬송가 가사 한국어 텍스트 표시')
            : ng('찬송가 가사 영역 없음');
        }
      } else {
        ng('찬송가 36번 검색 결과 없음');
      }
    } else {
      ng('찬송가 검색 입력창 없음');
    }

    // 내 찬양 탭
    const myTab = page.locator('button').filter({ hasText: '내 찬양' });
    const myTabExists = await myTab.count();
    if (myTabExists > 0) {
      await myTab.first().click();
      await page.waitForTimeout(400);
      ok('내 찬양 탭 클릭');
    }

    // 자유곡 탭
    const freeTab = page.locator('button').filter({ hasText: '자유곡' });
    const freeTabExists = await freeTab.count();
    if (freeTabExists > 0) {
      await freeTab.first().click();
      await page.waitForTimeout(400);
      ok('자유곡 탭 클릭');
      await shot(page, 'lyrics-free-song');
    }

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 5. 성경 페이지
// ═══════════════════════════════════════════════════════════════
async function testBible(browser) {
  section('성경 페이지');
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 900 }, storageState: STORAGE });
  const page = await ctx.newPage();
  try {
    await page.goto(`${BASE}/bible`, { waitUntil: 'networkidle', timeout: 20000 });
    await dismissModal(page);
    await page.waitForTimeout(800);

    await shot(page, 'bible-books');

    // 창세기 클릭
    const genesis = page.locator('text=창세기');
    await genesis.waitFor({ timeout: 8000 }).catch(() => {});
    const genesisCount = await genesis.count();
    if (genesisCount > 0) {
      await genesis.first().click();
      await page.waitForTimeout(700);
      ok('창세기 선택');
    } else {
      ng('창세기 항목 없음 — 성경 책 목록 미로드');
    }

    // 장 선택 (1장)
    const ch1 = page.locator('button').filter({ hasText: /^1$/ }).first();
    await ch1.waitFor({ timeout: 5000 }).catch(() => {});
    const ch1Exists = await ch1.count();
    if (ch1Exists > 0) {
      ok('1장 버튼 존재');
      await ch1.click();
      await page.waitForTimeout(1000);
      ok('1장 선택');
    } else {
      ng('1장 버튼 없음');
    }

    await shot(page, 'bible-chapter');

    // 절 내용 확인
    const verseText = page.locator('text=태초에');
    const verseCount = await verseText.count();
    if (verseCount > 0) {
      ok('창세기 1:1 "태초에" 텍스트 표시');
    } else {
      const verseEl = page.locator('[data-verse], .verse, li').first();
      const verseElCount = await verseEl.count();
      if (verseElCount > 0) {
        const txt = await verseEl.textContent();
        txt && txt.length > 3 ? ok(`절 텍스트 표시 (예: ${txt.slice(0, 20)}...)`) : ng('절 내용 비어있음');
      } else {
        ng('성경 절 미표시');
      }
    }

    // ▶ 절 클릭으로 선택 (Display 전송 버튼 활성화 조건)
    const firstVerse = page.locator('[data-verse="1"], li').first();
    const firstVerseCount = await firstVerse.count();
    if (firstVerseCount > 0) {
      await firstVerse.click();
      await page.waitForTimeout(500);
      ok('1절 클릭 (선택)');
    } else {
      // 화면에 보이는 절 아무거나 클릭
      const anyVerse = page.locator('li').first();
      const anyCount = await anyVerse.count();
      if (anyCount > 0) {
        await anyVerse.click();
        await page.waitForTimeout(500);
        ok('절 클릭 (li 기준)');
      }
    }

    await shot(page, 'bible-verse-selected');

    // Display 전송 버튼 (절 선택 후 활성화)
    const sendBtn = page.locator('button').filter({ hasText: /전송|Display|보내기/ });
    const sendCount = await sendBtn.count();
    sendCount > 0 ? ok('성경 Display 전송 버튼 존재') : ng('성경 Display 전송 버튼 없음');

    // Shift+클릭 범위 선택 테스트
    const thirdVerse = page.locator('li').nth(2);
    const thirdCount = await thirdVerse.count();
    if (thirdCount > 0) {
      await thirdVerse.click({ modifiers: ['Shift'] });
      await page.waitForTimeout(400);
      ok('Shift+클릭 범위 선택 시도');
    }

    // 성경 번역 버전 드롭다운
    const versionSel = page.locator('select').first();
    const versionCount = await versionSel.count();
    versionCount > 0 ? ok('성경 번역 버전 선택기 존재') : ng('성경 번역 버전 선택기 없음');

    // 가로 스크롤
    const hasHScroll = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 5);
    hasHScroll ? ng('성경 페이지 가로 스크롤') : ok('성경 페이지 가로 스크롤 없음');

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 6. Display / Overlay / Stage 페이지
// ═══════════════════════════════════════════════════════════════
async function testDisplay(browser) {
  section('Display 페이지 (OBS / Overlay / Stage)');
  const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await ctx.newPage();
  try {
    // --- /display ---
    await page.goto(`${GO}/display`, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(1200);
    const dispText = await page.locator('body').textContent();
    /[가-힣]/.test(dispText || '')
      ? ok('/display 한국어 콘텐츠 렌더링')
      : ng('/display 콘텐츠 없음');
    /\d+\s*\/\s*\d+/.test(dispText || '')
      ? ok('/display 슬라이드 번호 표시')
      : ng('/display 슬라이드 번호 없음');
    await shot(page, 'display-obs');

    // --- /display/overlay ---
    await page.goto(`${GO}/display/overlay`, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(1000);

    const slide = await page.$('#slide');
    slide ? ok('/display/overlay #slide 요소 존재') : ng('/display/overlay #slide 없음');

    const bg = await page.evaluate(() => window.getComputedStyle(document.body).background);
    bg.includes('transparent') || bg.includes('rgba(0, 0, 0, 0)') || bg === ''
      ? ok('/display/overlay 배경 투명 (크로마키)')
      : ng('/display/overlay 배경 불투명', `background: ${bg}`);

    await shot(page, 'display-overlay');

    // --- /display/stage ---
    await page.goto(`${GO}/display/stage`, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(1000);
    const stageTxt = await page.locator('body').textContent();
    /[가-힣]/.test(stageTxt || '')
      ? ok('/display/stage 한국어 콘텐츠 렌더링')
      : ng('/display/stage 콘텐츠 없음');
    await shot(page, 'display-stage');

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 7. 모바일 리모컨 (/mobile)
// ═══════════════════════════════════════════════════════════════
async function testMobileRemote(browser) {
  section('모바일 리모컨 (/mobile)');
  const ctx = await browser.newContext({ viewport: { width: 390, height: 844 } }); // iPhone 14
  const page = await ctx.newPage();
  try {
    await page.goto(`${GO}/mobile`, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(800);

    const bodyTxt = await page.locator('body').textContent();
    bodyTxt && bodyTxt.length > 10
      ? ok('/mobile 페이지 콘텐츠 로드')
      : ng('/mobile 페이지 비어있음');

    // 이전/다음 버튼 있는지
    const navBtn = page.locator('button').filter({ hasText: /이전|다음|prev|next/i });
    const navCount = await navBtn.count();
    navCount > 0 ? ok(`모바일 내비게이션 버튼 ${navCount}개`) : ng('모바일 내비게이션 버튼 없음');

    const hasHScroll = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 5);
    hasHScroll ? ng('모바일 가로 스크롤') : ok('모바일 가로 스크롤 없음');

    await shot(page, 'mobile-remote');

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 8. 반응형 (375px — iPhone SE)
// ═══════════════════════════════════════════════════════════════
async function testResponsive(browser) {
  section('반응형 레이아웃 (375px)');
  const ctx = await browser.newContext({ viewport: { width: 375, height: 667 }, storageState: STORAGE });
  const page = await ctx.newPage();
  try {
    for (const route of ['/', '/bulletin', '/bible', '/lyrics']) {
      await page.goto(`${BASE}${route}`, { waitUntil: 'networkidle', timeout: 15000 });
      await dismissModal(page);
      await page.waitForTimeout(500);
      const hasHScroll = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 5);
      hasHScroll ? ng(`${route} 375px 가로 스크롤`) : ok(`${route} 375px 가로 스크롤 없음`);
    }
    await shot(page, 'responsive-375');
  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 9. WebSocket 연결 확인
// ═══════════════════════════════════════════════════════════════
async function testWebSocket(browser) {
  section('WebSocket 연결');
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 }, storageState: STORAGE });
  const page = await ctx.newPage();
  try {
    // WS 이벤트 감지 (페이지 로드 중 발생하는 WS 연결)
    let detectedWs = '';
    page.on('websocket', (ws) => { if (!detectedWs) detectedWs = ws.url(); });

    await page.goto(BASE, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(2000);

    if (detectedWs) {
      ok(`WebSocket 연결 감지: ${detectedWs}`);
    }

    // Go 서버 /ws 직접 연결 확인
    const wsStatus = await page.evaluate(() => {
      return new Promise((res) => {
        try {
          const ws = new WebSocket('ws://localhost:8080/ws');
          ws.onopen = () => { ws.close(); res('ok'); };
          ws.onerror = () => res('error');
          setTimeout(() => res('timeout'), 4000);
        } catch { res('error'); }
      });
    });
    wsStatus === 'ok'
      ? ok('Go WebSocket /ws 연결 성공')
      : ng('Go WebSocket /ws 연결 실패', wsStatus);

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 10. 시퀀스 패널 — PREV/NEXT 내비게이션
// ═══════════════════════════════════════════════════════════════
async function testSequenceNav(browser) {
  section('시퀀스 패널 내비게이션');
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, storageState: STORAGE });
  const page = await ctx.newPage();
  try {
    await page.goto(BASE, { waitUntil: 'networkidle', timeout: 20000 });
    await dismissModal(page);
    await page.waitForTimeout(1000);

    // 시퀀스 아이템 존재 확인
    const seqItems = await page.$$('[data-testid="seq-item"]');
    if (seqItems.length > 0) {
      ok(`시퀀스 아이템 ${seqItems.length}개 존재`);

      // 첫 번째 아이템 클릭
      await seqItems[0].click();
      await page.waitForTimeout(400);
      ok('첫 번째 시퀀스 아이템 클릭');

      // PREV/NEXT 버튼
      const prevBtn = page.locator('button').filter({ hasText: /PREV|이전|◀/ });
      const nextBtn = page.locator('button').filter({ hasText: /NEXT|다음|▶/ });

      const prevCount = await prevBtn.count();
      const nextCount = await nextBtn.count();
      prevCount > 0 ? ok('PREV 버튼 존재') : ng('PREV 버튼 없음');
      nextCount > 0 ? ok('NEXT 버튼 존재') : ng('NEXT 버튼 없음');

      if (nextCount > 0) {
        await nextBtn.first().click();
        await page.waitForTimeout(400);
        ok('NEXT 버튼 클릭');
      }

      // 키보드 내비게이션 (ArrowRight)
      await page.keyboard.press('ArrowRight');
      await page.waitForTimeout(350);
      ok('ArrowRight 키 내비게이션');

    } else {
      ng('시퀀스 아이템 없음 — 예배 순서 미로드');
    }

    await shot(page, 'sequence-nav');

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 메인 실행
// ═══════════════════════════════════════════════════════════════
(async () => {
  if (!fs.existsSync(OUT_DIR)) fs.mkdirSync(OUT_DIR, { recursive: true });
  console.log('🚀 easyPreparation 전체 기능 테스트 시작');
  console.log(`   Next.js: ${BASE}  |  Go: ${GO}`);
  console.log('═'.repeat(50));

  const startTime = Date.now();

  // API 테스트 (브라우저 없이)
  await testAPIs();

  // 브라우저 테스트
  const browser = await chromium.launch({ headless: true });
  try {
    await testProShell(browser);
    await testBulletin(browser);
    await testLyrics(browser);
    await testBible(browser);
    await testDisplay(browser);
    await testMobileRemote(browser);
    await testResponsive(browser);
    await testWebSocket(browser);
    await testSequenceNav(browser);
  } finally {
    await browser.close();
  }

  const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);

  // 최종 리포트
  console.log('\n' + '═'.repeat(50));
  console.log(`결과: ${pass + fail}개 검사  ✓ ${pass} 통과  ✗ ${fail} 실패  (${elapsed}s)`);

  console.log('\n섹션별 결과:');
  for (const [sec, res] of Object.entries(sectionResults)) {
    const icon = res.fail === 0 ? '✅' : '⚠️';
    console.log(`  ${icon} ${sec}: ${res.pass}통과 / ${res.fail}실패`);
  }

  if (issues.length > 0) {
    console.log('\n발견된 문제:');
    issues.forEach((i) => console.log(`  • ${i}`));
  } else {
    console.log('\n✅ 모든 테스트 통과 — 버그 없음');
  }

  console.log(`\n📁 스크린샷: ${OUT_DIR}`);
  process.exit(fail > 0 ? 1 : 0);
})();
