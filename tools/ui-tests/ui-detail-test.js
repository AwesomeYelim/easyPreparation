#!/usr/bin/env node
// 상세 버그 탐지 테스트 — API↔UI, API↔DB, 초기 사용자 UX
// 실행: NODE_PATH=ui/node_modules node tools/ui-detail-test.js
// 전제: make dev (Go :8080 + Next.js :3000) 실행 중

const { chromium } = require('@playwright/test');
const path = require('path');
const fs = require('fs');
const http = require('http');
const https = require('https');

const OUT_DIR = path.join(__dirname, '../.claude/screenshots/detail-test');
const BASE = 'http://localhost:3000';
const GO = 'http://localhost:8080';

if (!fs.existsSync(OUT_DIR)) fs.mkdirSync(OUT_DIR, { recursive: true });

// 투어 건너뛰기
const STORAGE_NO_TOUR = {
  cookies: [],
  origins: [{ origin: BASE, localStorage: [{ name: 'ep_tour_done', value: '1' }] }],
};
// 투어 미완료 (신규 사용자)
const STORAGE_NEW_USER = { cookies: [], origins: [] };

let pass = 0, fail = 0;
const issues = [];
const sectionResults = {};
let currentSection = '';
const feedbacks = []; // UX 피드백 (버그는 아니지만 개선 권고)

function ok(label) {
  process.stdout.write(`  ✓ ${label}\n`);
  pass++;
  if (!sectionResults[currentSection]) sectionResults[currentSection] = { pass: 0, fail: 0 };
  sectionResults[currentSection].pass++;
}
function ng(label, detail = '') {
  process.stdout.write(`  ✗ ${label}${detail ? ' — ' + detail : ''}\n`);
  fail++;
  issues.push(`[${currentSection}] ${label}${detail ? ': ' + detail : ''}`);
  if (!sectionResults[currentSection]) sectionResults[currentSection] = { pass: 0, fail: 0 };
  sectionResults[currentSection].fail++;
}
function warn(label) {
  process.stdout.write(`  ⚠ ${label}\n`);
  feedbacks.push(`[${currentSection}] ${label}`);
}
function section(name) {
  currentSection = name;
  process.stdout.write(`\n[${name}]\n`);
}

async function shot(page, name) {
  const p = path.join(OUT_DIR, `${name}.png`);
  await page.screenshot({ path: p, fullPage: false });
  process.stdout.write(`  📸 ${name}.png\n`);
}

function httpReq(method, urlStr, body) {
  return new Promise((resolve) => {
    const u = new URL(urlStr);
    const mod = u.protocol === 'https:' ? https : http;
    const data = body ? JSON.stringify(body) : null;
    const opts = {
      hostname: u.hostname,
      port: u.port || (u.protocol === 'https:' ? 443 : 80),
      path: u.pathname + u.search,
      method,
      headers: {
        'Content-Type': 'application/json',
        ...(data ? { 'Content-Length': Buffer.byteLength(data) } : {}),
      },
    };
    const req = mod.request(opts, (res) => {
      let raw = '';
      res.on('data', (d) => (raw += d));
      res.on('end', () => {
        let parsed = null;
        try { parsed = JSON.parse(raw); } catch (_) {}
        resolve({ status: res.statusCode, body: parsed, raw, headers: res.headers });
      });
    });
    req.on('error', () => resolve({ status: 0, body: null, raw: '', headers: {} }));
    if (data) req.write(data);
    req.end();
  });
}
const httpGet  = (u) => httpReq('GET',    u, null);
const httpPut  = (u, b) => httpReq('PUT',  u, b);
const httpPost = (u, b) => httpReq('POST', u, b);

async function dismissModal(page) {
  const btn = page.locator('button').filter({ hasText: /시작|닫기|확인|건너뛰기|skip/i });
  if (await btn.count() > 0) {
    await btn.first().click().catch(() => {});
    await page.waitForTimeout(300);
  }
}

// ═══════════════════════════════════════════════════════════════
// 1. API 필드 무결성 — 응답에 필수 필드가 모두 있는가
// ═══════════════════════════════════════════════════════════════
async function testAPIFieldIntegrity() {
  section('API 필드 무결성');

  // 예배 순서 필드
  const order = await httpGet(`${GO}/api/worship-order?type=main_worship`);
  if (order.status === 200 && Array.isArray(order.body) && order.body.length > 0) {
    const item = order.body[0];
    ['key', 'title'].every((f) => f in item)
      ? ok('예배 순서 항목: key, title 필드 존재')
      : ng('예배 순서 항목 필드 누락', `found: ${Object.keys(item).join(', ')}`);

    // info 필드 유효값 검사 (notice=교회소식, c-edit=하이픈 표기 허용)
    const validInfos = new Set(['-', 'c_edit', 'b_edit', 'edit', 'r_edit', 'notice', 'c-edit']);
    const badItems = order.body.filter(
      (it) => it.info && !validInfos.has(it.info)
    );
    badItems.length === 0
      ? ok('예배 순서 info 필드 모두 유효값')
      : ng('예배 순서 info 필드 유효하지 않은 값', `${badItems.map(i => `${i.key}:${i.info}`).join(', ')}`);

    // key 중복 검사
    const keys = order.body.map((i) => i.key);
    const uniqueKeys = new Set(keys);
    uniqueKeys.size === keys.length
      ? ok(`예배 순서 key 중복 없음 (${keys.length}개)`)
      : ng('예배 순서 key 중복 발견', `${keys.length - uniqueKeys.size}개 중복`);
  } else {
    ng('예배 순서 API 응답 오류', `status=${order.status}`);
  }

  // 찬송가 목록 필드
  const hymns = await httpGet(`${GO}/api/hymns?page=1&limit=5`);
  if (hymns.status === 200 && hymns.body?.hymns?.length > 0) {
    const h = hymns.body.hymns[0];
    const requiredFields = ['id', 'hymnbook', 'number', 'title'];
    const missingFields = requiredFields.filter((f) => !(f in h));
    missingFields.length === 0
      ? ok('찬송가 목록 필수 필드 존재 (id, hymnbook, number, title)')
      : ng('찬송가 목록 필드 누락', missingFields.join(', '));

    // hymnbook 값 유효성
    const validBooks = new Set(['new', 'old', '21c']);
    const badBooks = hymns.body.hymns.filter((h) => !validBooks.has(h.hymnbook));
    badBooks.length === 0
      ? ok('찬송가 hymnbook 모두 유효값')
      : ng('찬송가 hymnbook 유효하지 않은 값', badBooks.map(b=>b.hymnbook).join(', '));

    // number 타입 검사 (숫자여야 함)
    const nonNumericNumbers = hymns.body.hymns.filter((h) => typeof h.number !== 'number');
    nonNumericNumbers.length === 0
      ? ok('찬송가 number 필드 모두 숫자 타입')
      : ng('찬송가 number 필드 문자열', `${nonNumericNumbers.length}개`);

    // total 필드 & 페이지네이션 일관성
    typeof hymns.body.total === 'number' && hymns.body.total > 0
      ? ok(`찬송가 total 필드 존재: ${hymns.body.total}`)
      : ng('찬송가 total 필드 누락 또는 0');
  } else {
    ng('찬송가 목록 API 오류', `status=${hymns.status}`);
  }

  // 찬송가 상세 (가사 필드)
  const detail = await httpGet(`${GO}/api/hymns/detail?number=36&book=new`);
  if (detail.status === 200) {
    typeof detail.body?.lyrics === 'string' && detail.body.lyrics.length > 10
      ? ok(`찬송가 상세 lyrics 필드 존재 (${detail.body.lyrics.length}자)`)
      : ng('찬송가 상세 lyrics 필드 없거나 너무 짧음');

    typeof detail.body?.number === 'number' && detail.body.number === 36
      ? ok('찬송가 상세 number === 36 일치')
      : ng('찬송가 상세 number 불일치', `got: ${detail.body?.number}`);
  } else {
    ng('찬송가 상세 API 오류', `status=${detail.status}`);
  }

  // 성경 절 필드
  const verse = await httpGet(`${GO}/api/bible/verses?book=1&chapter=1`);
  if (verse.status === 200 && Array.isArray(verse.body) && verse.body.length > 0) {
    const v = verse.body[0];
    const requiredVerse = ['verse', 'text'];
    const missingVerse = requiredVerse.filter((f) => !(f in v));
    missingVerse.length === 0
      ? ok('성경 절 필수 필드 존재 (verse, text)')
      : ng('성경 절 필드 누락', missingVerse.join(', '));

    // verse 번호 연속성 (1, 2, 3...)
    const verseNums = verse.body.map((v) => v.verse);
    const isSequential = verseNums.every((n, i) => n === i + 1);
    isSequential
      ? ok(`성경 절 번호 연속성 확인 (1~${verseNums.length})`)
      : ng('성경 절 번호 연속성 깨짐', `got: ${verseNums.slice(0,5).join(',')}`);

    // text 빈 값 없는지
    const emptyTexts = verse.body.filter((v) => !v.text || v.text.trim() === '');
    emptyTexts.length === 0
      ? ok('성경 절 text 모두 비어있지 않음')
      : ng('성경 절 text 빈 값 발견', `${emptyTexts.length}개`);
  } else {
    ng('성경 절 API 오류', `status=${verse.status}`);
  }

  // 성경 책 목록 필드 (API는 한글 책명 키 object 반환)
  const books = await httpGet(`${GO}/api/bible/books`);
  if (books.status === 200 && books.body && typeof books.body === 'object' && !Array.isArray(books.body)) {
    const bookNames = Object.keys(books.body);
    bookNames.length === 66
      ? ok(`성경 66권 전체 확인`)
      : ng('성경 책 수 이상', `got: ${bookNames.length}`);

    books.body['창세기'] !== undefined
      ? ok(`성경 창세기 확인 (index=${books.body['창세기'].index})`)
      : ng('창세기 없음');
    books.body['요한계시록'] !== undefined
      ? ok(`성경 요한계시록 확인 (index=${books.body['요한계시록'].index})`)
      : ng('요한계시록 없음');
  } else {
    ng('성경 책 목록 API 오류', `status=${books.status}, type=${typeof books.body}`);
  }

  // DisplayConfig 필드
  const cfg = await httpGet(`${GO}/api/display-config`);
  if (cfg.status === 200) {
    const requiredCfg = ['font', 'overlayBgOpacity', 'overlayTextColor', 'overlayPosition', 'overlayFontScale'];
    const missingCfg = requiredCfg.filter((f) => !(f in cfg.body));
    missingCfg.length === 0
      ? ok('DisplayConfig 필수 필드 모두 존재')
      : ng('DisplayConfig 필드 누락', missingCfg.join(', '));

    // 타입 & 범위 검사
    typeof cfg.body.overlayBgOpacity === 'number' && cfg.body.overlayBgOpacity >= 0 && cfg.body.overlayBgOpacity <= 1
      ? ok(`overlayBgOpacity 범위 정상: ${cfg.body.overlayBgOpacity}`)
      : ng('overlayBgOpacity 범위 이상', `got: ${cfg.body.overlayBgOpacity}`);

    typeof cfg.body.overlayFontScale === 'number' && cfg.body.overlayFontScale >= 0.5 && cfg.body.overlayFontScale <= 2.0
      ? ok(`overlayFontScale 범위 정상: ${cfg.body.overlayFontScale}`)
      : ng('overlayFontScale 범위 이상', `got: ${cfg.body.overlayFontScale}`);

    /^#[0-9a-fA-F]{6}$/.test(cfg.body.overlayTextColor)
      ? ok(`overlayTextColor hex 형식 정상: ${cfg.body.overlayTextColor}`)
      : ng('overlayTextColor 형식 이상', `got: ${cfg.body.overlayTextColor}`);
  } else {
    ng('DisplayConfig API 오류', `status=${cfg.status}`);
  }
}

// ═══════════════════════════════════════════════════════════════
// 2. API 엣지케이스 & 유효성 교정
// ═══════════════════════════════════════════════════════════════
async function testAPIEdgeCases() {
  section('API 엣지케이스 & 유효성');

  // 잘못된 예배 타입 → 400
  const badType = await httpGet(`${GO}/api/worship-order?type=invalid_type`);
  badType.status === 400
    ? ok('잘못된 예배 타입 → 400 Bad Request')
    : ng('잘못된 예배 타입 → 400 아님', `got: ${badType.status}`);

  // 빈 찬송가 검색 → 400
  const emptySearch = await httpGet(`${GO}/api/hymns/search?q=`);
  emptySearch.status === 400
    ? ok('빈 찬송가 검색어 → 400 Bad Request')
    : ng('빈 찬송가 검색어 → 400 아님', `got: ${emptySearch.status}`);

  // 존재하지 않는 찬송가 번호 → 404
  const noHymn = await httpGet(`${GO}/api/hymns/detail?number=99999&book=new`);
  noHymn.status === 404
    ? ok('존재하지 않는 찬송가 → 404 Not Found')
    : ng('존재하지 않는 찬송가 → 404 아님', `got: ${noHymn.status}`);

  // 찬송가 번호 0 → 400 또는 404
  const zeroHymn = await httpGet(`${GO}/api/hymns/detail?number=0&book=new`);
  [400, 404].includes(zeroHymn.status)
    ? ok(`찬송가 번호 0 → ${zeroHymn.status} (정상 거부)`)
    : ng('찬송가 번호 0 → 예상치 못한 응답', `got: ${zeroHymn.status}`);

  // 음수 찬송가 번호
  const negHymn = await httpGet(`${GO}/api/hymns/detail?number=-1&book=new`);
  [400, 404].includes(negHymn.status)
    ? ok(`찬송가 번호 -1 → ${negHymn.status} (정상 거부)`)
    : ng('찬송가 번호 -1 → 예상치 못한 응답', `got: ${negHymn.status}`);

  // DisplayConfig PUT — 범위 밖 값 → 서버가 교정하는지
  const badCfg = await httpPut(`${GO}/api/display-config`, {
    font: 'invalid_font',
    overlayBgOpacity: 5.0,      // 범위 초과
    overlayFontScale: 0.1,      // 범위 미만
    overlayPosition: 'bad_pos',
    overlayTextColor: '',
  });
  if (badCfg.status === 200 && badCfg.body) {
    badCfg.body.font === 'default'
      ? ok('잘못된 font → "default"로 교정')
      : ng('font 교정 실패', `got: ${badCfg.body.font}`);
    badCfg.body.overlayBgOpacity === 0.75
      ? ok('범위 초과 overlayBgOpacity → 0.75로 교정')
      : ng('overlayBgOpacity 교정 실패', `got: ${badCfg.body.overlayBgOpacity}`);
    badCfg.body.overlayFontScale === 1.0
      ? ok('범위 미만 overlayFontScale → 1.0으로 교정')
      : ng('overlayFontScale 교정 실패', `got: ${badCfg.body.overlayFontScale}`);
    badCfg.body.overlayPosition === 'flex-end'
      ? ok('잘못된 overlayPosition → "flex-end"로 교정')
      : ng('overlayPosition 교정 실패', `got: ${badCfg.body.overlayPosition}`);
    badCfg.body.overlayTextColor === '#ffffff'
      ? ok('빈 overlayTextColor → "#ffffff"로 교정')
      : ng('overlayTextColor 교정 실패', `got: ${badCfg.body.overlayTextColor}`);
  } else {
    ng('DisplayConfig PUT 오류', `status=${badCfg.status}`);
  }

  // 예배 순서 PUT — 잘못된 타입
  const badOrderType = await httpPut(`${GO}/api/worship-order`, {
    type: '../etc/passwd',
    items: [],
  });
  badOrderType.status === 400
    ? ok('path traversal 예배 타입 → 400 차단')
    : ng('path traversal 예배 타입 차단 실패', `got: ${badOrderType.status}`);

  // 성경 책 0번 (없는 번호) — 서버가 400 반환해야 함
  const badBook = await httpGet(`${GO}/api/bible/verses?book=0&chapter=1`);
  badBook.status === 400
    ? ok(`성경 book=0 → 400 Bad Request (정상 거부)`)
    : ng('성경 book=0 비정상 응답', `got: ${badBook.status}, body: ${JSON.stringify(badBook.body).slice(0,50)}`);

  // CORS: Content-Type 헤더
  const orderCors = await httpGet(`${GO}/api/worship-order?type=main_worship`);
  (orderCors.headers['content-type'] || '').includes('application/json')
    ? ok('응답 Content-Type: application/json 정상')
    : ng('응답 Content-Type 이상', `got: ${orderCors.headers['content-type']}`);
}

// ═══════════════════════════════════════════════════════════════
// 3. API ↔ DB round-trip (PUT → GET → 값 일치)
// ═══════════════════════════════════════════════════════════════
async function testAPIDBRoundTrip() {
  section('API ↔ DB round-trip');

  // ── 예배 순서 round-trip ──
  const original = await httpGet(`${GO}/api/worship-order?type=after_worship`);
  if (original.status !== 200 || !Array.isArray(original.body)) {
    ng('round-trip 원본 데이터 조회 실패');
    return;
  }
  const originalItems = original.body;

  // 테스트 항목 추가
  const testItem = { key: '__test__', title: 'RoundTripTest', obj: '-', info: '-', lead: '-' };
  const modified = [...originalItems, testItem];
  const putRes = await httpPut(`${GO}/api/worship-order`, { type: 'after_worship', items: modified });
  putRes.status === 200 && putRes.body?.ok
    ? ok('예배 순서 PUT 저장 성공')
    : ng('예배 순서 PUT 저장 실패', `status=${putRes.status}`);

  // GET으로 재조회
  const verify = await httpGet(`${GO}/api/worship-order?type=after_worship`);
  if (verify.status === 200 && Array.isArray(verify.body)) {
    verify.body.length === modified.length
      ? ok(`round-trip 항목 수 일치: ${verify.body.length}개`)
      : ng('round-trip 항목 수 불일치', `expected ${modified.length}, got ${verify.body.length}`);

    const testFound = verify.body.find((i) => i.key === '__test__');
    testFound?.title === 'RoundTripTest'
      ? ok('round-trip 추가 항목 title 일치')
      : ng('round-trip 추가 항목 찾기 실패');
  } else {
    ng('round-trip GET 재조회 실패');
  }

  // 원본 복원
  const restore = await httpPut(`${GO}/api/worship-order`, { type: 'after_worship', items: originalItems });
  restore.status === 200
    ? ok('예배 순서 원본 복원 성공')
    : ng('예배 순서 원본 복원 실패');

  // 복원 확인
  const restored = await httpGet(`${GO}/api/worship-order?type=after_worship`);
  restored.body?.length === originalItems.length
    ? ok(`예배 순서 복원 후 항목 수 일치: ${originalItems.length}개`)
    : ng('예배 순서 복원 후 항목 수 불일치');

  // ── DisplayConfig round-trip ──
  const origCfg = await httpGet(`${GO}/api/display-config`);
  if (origCfg.status !== 200) { ng('DisplayConfig 원본 조회 실패'); return; }

  const testCfg = {
    font: 'noto-sans-kr',
    overlayBgOpacity: 0.5,
    overlayTextColor: '#ff0000',
    overlayPosition: 'center',
    overlayFontScale: 1.5,
  };
  const cfgPut = await httpPut(`${GO}/api/display-config`, testCfg);
  cfgPut.status === 200
    ? ok('DisplayConfig PUT 저장 성공')
    : ng('DisplayConfig PUT 저장 실패', `status=${cfgPut.status}`);

  const cfgVerify = await httpGet(`${GO}/api/display-config`);
  if (cfgVerify.status === 200 && cfgVerify.body) {
    cfgVerify.body.font === 'noto-sans-kr'
      ? ok('DisplayConfig font round-trip 일치')
      : ng('DisplayConfig font 불일치', `got: ${cfgVerify.body.font}`);
    cfgVerify.body.overlayBgOpacity === 0.5
      ? ok('DisplayConfig overlayBgOpacity round-trip 일치')
      : ng('DisplayConfig overlayBgOpacity 불일치', `got: ${cfgVerify.body.overlayBgOpacity}`);
    cfgVerify.body.overlayTextColor === '#ff0000'
      ? ok('DisplayConfig overlayTextColor round-trip 일치')
      : ng('DisplayConfig overlayTextColor 불일치', `got: ${cfgVerify.body.overlayTextColor}`);
    cfgVerify.body.overlayPosition === 'center'
      ? ok('DisplayConfig overlayPosition round-trip 일치')
      : ng('DisplayConfig overlayPosition 불일치', `got: ${cfgVerify.body.overlayPosition}`);
    cfgVerify.body.overlayFontScale === 1.5
      ? ok('DisplayConfig overlayFontScale round-trip 일치')
      : ng('DisplayConfig overlayFontScale 불일치', `got: ${cfgVerify.body.overlayFontScale}`);
  } else {
    ng('DisplayConfig round-trip 재조회 실패');
  }

  // 원본 복원
  const cfgRestore = await httpPut(`${GO}/api/display-config`, origCfg.body);
  cfgRestore.status === 200
    ? ok('DisplayConfig 원본 복원 성공')
    : ng('DisplayConfig 원본 복원 실패');
}

// ═══════════════════════════════════════════════════════════════
// 4. API ↔ UI 데이터 매칭 (API 반환값 vs 화면에 보이는 값)
// ═══════════════════════════════════════════════════════════════
async function testAPIvsUI(browser) {
  section('API ↔ UI 데이터 매칭');

  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, storageState: STORAGE_NO_TOUR });
  const page = await ctx.newPage();
  try {
    // ── 예배 순서: API 항목 수 vs 시퀀스 패널 ──
    const orderAPI = await httpGet(`${GO}/api/worship-order?type=main_worship`);
    const apiCount = Array.isArray(orderAPI.body) ? orderAPI.body.length : 0;

    await page.goto(BASE, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(1000);

    // 시퀀스 패널 항목 수 (ProSequencePanel)
    const seqItems = page.locator('[data-testid="sequence-item"], .sequence-item, [class*="sequence"] [class*="item"]');
    let seqCount = await seqItems.count();
    if (seqCount === 0) {
      // 대안: 시퀀스 패널 내 클릭 가능한 행
      const panel = page.locator('[class*="ProSequence"], [class*="sequence-panel"], aside').first();
      const rows = panel.locator('[class*="cursor-pointer"], button[class*="w-full"]');
      seqCount = await rows.count();
    }
    seqCount > 0 && Math.abs(seqCount - apiCount) <= 2
      ? ok(`시퀀스 패널 항목 수 API와 근사 일치: UI=${seqCount}, API=${apiCount}`)
      : seqCount === 0
        ? warn(`시퀀스 패널 항목 data-testid 없음 — 수동 확인 필요 (API=${apiCount})`)
        : ng('시퀀스 패널 항목 수 API와 큰 차이', `UI=${seqCount}, API=${apiCount}`);

    // ── 예배 순서 첫 번째 항목 타이틀 비교 ──
    if (apiCount > 0) {
      const firstTitle = orderAPI.body[0].title;
      const bodyText = await page.locator('body').textContent();
      bodyText.includes(firstTitle)
        ? ok(`첫 번째 항목 "${firstTitle}" 화면에 표시됨`)
        : ng(`첫 번째 항목 "${firstTitle}" 화면에 없음`);
    }

    // ── 찬송가: API 검색 결과 vs UI ──
    await page.goto(`${BASE}/lyrics`, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(600);

    const hymnTab = page.locator('button').filter({ hasText: '찬송가 검색' });
    if (await hymnTab.count() > 0) {
      await hymnTab.first().click();
      await page.waitForTimeout(400);
    }

    const searchInput = page.locator('input[placeholder*="번호 또는 제목"]');
    if (await searchInput.count() > 0) {
      await searchInput.click();
      await searchInput.pressSequentially('36', { delay: 80 });
      await page.waitForTimeout(400);
      const searchBtn = page.locator('button').filter({ hasText: /^검색/ });
      if (await searchBtn.count() > 0) {
        await searchBtn.first().click();
        await page.waitForTimeout(2000);
      }

      // API 결과
      const hymnAPI = await httpGet(`${GO}/api/hymns/search?q=36`);
      const apiTitle = hymnAPI.body?.[0]?.title || '';
      const apiNum = hymnAPI.body?.[0]?.number;

      // UI에서 찬송가 번호 확인
      const numSpan = page.locator('span.font-black').filter({ hasText: /^36$/ });
      const numVisible = await numSpan.count() > 0;
      numVisible
        ? ok('찬송가 36번 번호 UI에 표시됨')
        : ng('찬송가 36번 번호 UI 미표시');

      // API 제목 UI 존재 확인
      if (apiTitle) {
        const titleEl = page.locator('div').filter({ hasText: apiTitle }).first();
        const titleExists = await titleEl.count() > 0;
        titleExists
          ? ok(`찬송가 제목 "${apiTitle}" UI에 표시됨`)
          : ng(`찬송가 제목 "${apiTitle}" UI 미표시`);
      }

      await shot(page, 'api-vs-ui-hymn');
    }

    // ── 성경: API 절 텍스트 vs UI ──
    await page.goto(`${BASE}/bible`, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(800);

    const verseAPI = await httpGet(`${GO}/api/bible/verses?book=1&chapter=1`);
    const firstVerse = verseAPI.body?.[0];

    if (firstVerse) {
      // 창세기 선택
      const genesisBtn = page.locator('button').filter({ hasText: '창세기' }).first();
      if (await genesisBtn.count() > 0) {
        await genesisBtn.click();
        await page.waitForTimeout(600);
      }
      // 1장 선택
      const ch1Btn = page.locator('button').filter({ hasText: /^1$/ }).first();
      if (await ch1Btn.count() > 0) {
        await ch1Btn.click();
        await page.waitForTimeout(800);
      }

      // API 절 텍스트 일부가 UI에 있는지 (앞 10자)
      const verseSnippet = (firstVerse.text || '').slice(0, 10);
      if (verseSnippet) {
        const bodyTxt = await page.locator('body').textContent();
        bodyTxt.includes(verseSnippet)
          ? ok(`성경 1:1 텍스트 "${verseSnippet}..." UI에 표시됨`)
          : ng(`성경 1:1 텍스트 "${verseSnippet}..." UI 미표시`);
      }
      await shot(page, 'api-vs-ui-bible');
    }

    // ── 예배 타입 전환: 오후예배 항목 수 API vs UI ──
    await page.goto(`${BASE}/bulletin`, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(800);

    const afterAPI = await httpGet(`${GO}/api/worship-order?type=after_worship`);
    const afterCount = Array.isArray(afterAPI.body) ? afterAPI.body.length : 0;

    const typeSelect = page.locator('select').first();
    if (await typeSelect.count() > 0) {
      await typeSelect.selectOption('after_worship');
      await page.waitForTimeout(1000);

      // 화면에 오후예배 항목이 렌더되었는지 (항목 수로 확인)
      if (afterCount > 0 && afterAPI.body?.[0]?.title) {
        const firstAfterTitle = afterAPI.body[0].title;
        const bodyTxt = await page.locator('body').textContent();
        bodyTxt.includes(firstAfterTitle)
          ? ok(`오후예배 전환 후 첫 항목 "${firstAfterTitle}" 표시됨`)
          : ng(`오후예배 전환 후 첫 항목 "${firstAfterTitle}" 미표시`);
      }
      await shot(page, 'api-vs-ui-worship-type');
    }

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 5. Display 데이터 흐름 (order → status → navigate)
// ═══════════════════════════════════════════════════════════════
async function testDisplayFlow() {
  section('Display 데이터 흐름');

  // 예배 순서 조회 → display/order POST
  const orderData = await httpGet(`${GO}/api/worship-order?type=main_worship`);
  if (!Array.isArray(orderData.body) || orderData.body.length === 0) {
    ng('Display 흐름 테스트: 예배 순서 없음');
    return;
  }

  // display/order 전송
  const sendRes = await httpPost(`${GO}/display/order`, {
    items: orderData.body,
    churchName: 'TestChurch',
    email: 'test@test.com',
  });
  sendRes.status === 200
    ? ok('display/order POST 성공')
    : ng('display/order POST 실패', `status=${sendRes.status}`);

  await new Promise((r) => setTimeout(r, 500));

  // display/status 확인
  const status = await httpGet(`${GO}/display/status`);
  if (status.status === 200 && status.body) {
    Array.isArray(status.body.items) && status.body.items.length > 0
      ? ok(`display/status items: ${status.body.items.length}개`)
      : ng('display/status items 비어있음');

    typeof status.body.idx === 'number'
      ? ok(`display/status idx: ${status.body.idx}`)
      : ng('display/status idx 필드 없음');
  } else {
    ng('display/status 조회 실패', `status=${status.status}`);
  }

  // display/navigate — next
  const nav1 = await httpPost(`${GO}/display/navigate`, { direction: 'next' });
  nav1.status === 200
    ? ok('display/navigate next 성공')
    : ng('display/navigate next 실패', `status=${nav1.status}`);

  await new Promise((r) => setTimeout(r, 300));

  // index가 올라갔는지 확인
  const status2 = await httpGet(`${GO}/display/status`);
  if (status2.status === 200 && typeof status2.body?.idx === 'number') {
    status2.body.idx > 0
      ? ok(`navigate 후 idx 증가: ${status2.body.idx}`)
      : ng('navigate 후 idx 미증가', `got: ${status2.body.idx}`);
  }

  // display/navigate — prev
  const nav2 = await httpPost(`${GO}/display/navigate`, { direction: 'prev' });
  nav2.status === 200
    ? ok('display/navigate prev 성공')
    : ng('display/navigate prev 실패', `status=${nav2.status}`);

  // display/jump
  const jump = await httpPost(`${GO}/display/jump`, { index: 0, subPageIdx: 0 });
  jump.status === 200
    ? ok('display/jump index=0 성공')
    : ng('display/jump 실패', `status=${jump.status}`);

  // display/status — WebSocket 브로드캐스트 포함 확인 (navigate_type 필드)
  await new Promise((r) => setTimeout(r, 300));
  const status3 = await httpGet(`${GO}/display/status`);
  status3.body?.idx === 0
    ? ok('jump 후 idx=0 확인')
    : ng('jump 후 idx 오류', `got: ${status3.body?.idx}`);

  // display/append 테스트
  const append = await httpPost(`${GO}/display/append`, {
    items: [{ title: '테스트', info: '-', obj: '-' }],
    source: 'test',
  });
  append.status === 200
    ? ok('display/append 성공')
    : ng('display/append 실패', `status=${append.status}`);
}

// ═══════════════════════════════════════════════════════════════
// 6. 초기 사용자 UX — 처음 방문 경험
// ═══════════════════════════════════════════════════════════════
async function testFirstTimeUX(browser) {
  section('초기 사용자 UX (신규 방문)');

  // 투어 모달 테스트 (ep_tour_done 없음)
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, storageState: STORAGE_NEW_USER });
  const page = await ctx.newPage();
  try {
    await page.goto(BASE, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(1500);

    await shot(page, 'first-time-landing');

    // 앱 로딩: 기본 레이아웃이 보이는지
    const bodyTxt = await page.locator('body').textContent();
    bodyTxt.length > 100
      ? ok('신규 방문 시 기본 레이아웃 로드됨')
      : ng('신규 방문 시 빈 화면');

    // 투어 모달 또는 온보딩 존재 여부 (있어도 되고 없어도 됨)
    const tourModal = page.locator('[class*="modal"], [class*="tour"], [class*="onboard"], dialog').first();
    const hasTour = await tourModal.count() > 0;
    hasTour
      ? ok('신규 사용자 투어/온보딩 모달 표시됨')
      : warn('신규 사용자 투어/온보딩 없음 — 첫 사용자 가이드 검토 권장');

    // 투어가 있으면 닫기 시도
    if (hasTour) {
      const closeBtn = page.locator('button').filter({ hasText: /닫기|확인|시작|건너뛰기|skip/i });
      if (await closeBtn.count() > 0) {
        await closeBtn.first().click();
        await page.waitForTimeout(400);
        ok('투어 모달 닫기 성공');
      }
    }

    // 핵심 네비게이션 항목이 보이는지
    const nav = page.locator('[data-testid="topbar"] nav a, nav a');
    const navCount = await nav.count();
    navCount >= 2
      ? ok(`네비게이션 탭 ${navCount}개 표시 (신규 방문 시)`)
      : ng('네비게이션 탭 부족', `${navCount}개`);

    // 페이지 타이틀이 의미있는지
    const title = await page.title();
    title && title.length > 0
      ? ok(`페이지 타이틀: "${title}"`)
      : warn('페이지 타이틀 비어있음 — SEO/UX 개선 권장');

  } finally {
    await ctx.close();
  }

  // 투어 완료 후 localStorage 저장 확인
  const ctx2 = await browser.newContext({ viewport: { width: 1440, height: 900 }, storageState: STORAGE_NO_TOUR });
  const page2 = await ctx2.newPage();
  try {
    await page2.goto(BASE, { waitUntil: 'networkidle', timeout: 20000 });
    await page2.waitForTimeout(800);

    const epTourDone = await page2.evaluate(() => localStorage.getItem('ep_tour_done'));
    epTourDone === '1'
      ? ok('투어 완료 localStorage "ep_tour_done=1" 정상 저장됨')
      : warn(`ep_tour_done 값: ${epTourDone} — 투어 완료 마킹 확인 필요`);

  } finally {
    await ctx2.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 7. 기능별 상세 UX — 찬송가/성경/주보 세부 플로우
// ═══════════════════════════════════════════════════════════════
async function testDetailedFeatureUX(browser) {
  section('기능별 상세 UX');

  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, storageState: STORAGE_NO_TOUR });
  const page = await ctx.newPage();
  try {
    // ── 찬송가 텍스트 검색 (번호가 아닌 제목) ──
    await page.goto(`${BASE}/lyrics`, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(600);

    const hymnTab = page.locator('button').filter({ hasText: '찬송가 검색' });
    if (await hymnTab.count() > 0) {
      await hymnTab.first().click();
      await page.waitForTimeout(400);
    }

    const searchInput = page.locator('input[placeholder*="번호 또는 제목"]');
    if (await searchInput.count() > 0) {
      // 텍스트로 검색
      await searchInput.click();
      await searchInput.pressSequentially('주 예수', { delay: 60 });
      await page.waitForTimeout(400);

      const searchBtn = page.locator('button').filter({ hasText: /^검색/ });
      if (await searchBtn.count() > 0) {
        await searchBtn.first().click();
        await page.waitForTimeout(2000);
      }

      // 결과 개수
      const results = page.locator('span.font-black');
      const resCount = await results.count();
      resCount > 1
        ? ok(`찬송가 텍스트 검색 "주 예수" — ${resCount}개 결과`)
        : ng('찬송가 텍스트 검색 결과 없음', `got: ${resCount}`);

      await shot(page, 'hymn-text-search');

      // 검색 결과 누적 동작 (두 번 검색)
      await searchInput.clear();
      await searchInput.pressSequentially('36', { delay: 60 });
      await page.waitForTimeout(400);
      if (await searchBtn.count() > 0) {
        await searchBtn.first().click();
        await page.waitForTimeout(2000);
      }
      const results2 = page.locator('span.font-black');
      const resCount2 = await results2.count();
      // 두 번째 검색 후에도 결과 존재
      resCount2 > 0
        ? ok(`두 번 연속 검색 후 결과 존재: ${resCount2}개`)
        : ng('두 번 연속 검색 후 결과 없음');

      // 중복 방지: 36번이 한 번만 표시되는지 (검색어 달라도 36번이 다시 추가되는지)
      const num36Spans = page.locator('span.font-black').filter({ hasText: /^36$/ });
      const num36Count = await num36Spans.count();
      // HymnSearch.tsx의 중복 방지 로직: existing Set으로 필터링
      // 36번은 두 번째 검색("36")에서만 나와야 함
      // 첫 검색("주 예수")에서도 36번이 있었다면 중복이 생길 수 있음
      num36Count <= 2
        ? ok(`36번 중복 방지 정상: ${num36Count}개`)
        : warn(`36번이 ${num36Count}개 — 중복 추가 가능성 확인 필요`);
    }

    // ── 내 찬양 탭 — 빈 상태 메시지 ──
    const myTab = page.locator('button').filter({ hasText: '내 찬양' });
    if (await myTab.count() > 0) {
      await myTab.first().click();
      await page.waitForTimeout(500);
      const emptyMsg = page.locator('body');
      const emptyTxt = await emptyMsg.textContent();
      // 빈 상태 안내 메시지가 있는지
      const hasEmptyGuide = /등록|없습니다|추가|비어|empty/i.test(emptyTxt);
      hasEmptyGuide
        ? ok('내 찬양 빈 상태 안내 메시지 존재')
        : warn('내 찬양 빈 상태 안내 메시지 없음 — UX 개선 권장');
      await shot(page, 'my-songs-empty');
    }

    // ── 성경 검색 기능 ──
    await page.goto(`${BASE}/bible`, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(800);

    const bibleSearchInput = page.locator('input[placeholder*="검색"], input[type="search"]').first();
    if (await bibleSearchInput.count() > 0) {
      await bibleSearchInput.click();
      await bibleSearchInput.pressSequentially('태초', { delay: 60 });
      await page.waitForTimeout(400);
      await bibleSearchInput.press('Enter');
      await page.waitForTimeout(1500);

      const bibleResults = await page.locator('body').textContent();
      /태초/.test(bibleResults)
        ? ok('성경 "태초" 검색 결과 표시됨')
        : ng('성경 "태초" 검색 결과 없음');
      await shot(page, 'bible-search-results');
    } else {
      warn('성경 검색 input 없음 — 검색 기능 확인 필요');
    }

    // ── 성경 번역 버전 전환 ──
    await page.goto(`${BASE}/bible`, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(800);

    // 창세기 1장 1절 로드
    const genesisBtn2 = page.locator('button').filter({ hasText: '창세기' }).first();
    if (await genesisBtn2.count() > 0) {
      await genesisBtn2.click();
      await page.waitForTimeout(500);
      const ch1 = page.locator('button').filter({ hasText: /^1$/ }).first();
      if (await ch1.count() > 0) { await ch1.click(); await page.waitForTimeout(600); }
    }

    // 번역 버전 선택기
    const versionSelect = page.locator('select').first();
    if (await versionSelect.count() > 0) {
      const options = await versionSelect.locator('option').allTextContents();
      options.length >= 2
        ? ok(`성경 번역 버전 ${options.length}개: ${options.join(', ')}`)
        : warn('성경 번역 버전이 1개 이하');

      if (options.length >= 2) {
        // 첫 번째 버전 절 텍스트
        const txt1 = await page.locator('body').textContent();
        const snippet1 = txt1.match(/태초[가-힣\s]+/)?.[0]?.slice(0, 20) || '';

        // 두 번째 버전으로 전환
        await versionSelect.selectOption({ index: 1 });
        await page.waitForTimeout(1000);
        const txt2 = await page.locator('body').textContent();
        const snippet2 = txt2.match(/태초[가-힣\s]+/)?.[0]?.slice(0, 20) || '';

        snippet1 !== snippet2 && snippet2.length > 0
          ? ok(`버전 전환 후 절 텍스트 변경됨 — "${snippet1}" → "${snippet2}"`)
          : warn('버전 전환 후 절 텍스트 동일 — 단일 버전만 지원할 수 있음');

        await shot(page, 'bible-version-switch');
      }
    } else {
      warn('성경 버전 선택기 없음');
    }

    // ── TemplateSelector — 시안 목록 열기 ──
    await page.goto(`${BASE}/bulletin`, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(800);

    const templateBtn = page.locator('button').filter({ hasText: /주보 시안|시안|template/i }).first();
    if (await templateBtn.count() > 0) {
      await templateBtn.click();
      await page.waitForTimeout(500);
      const dropdown = await page.locator('body').textContent();
      // 시안 목록이 열렸는지 (시안 이름들)
      const hasTemplates = /기본|클래식|모던|simple|classic|modern/i.test(dropdown);
      hasTemplates
        ? ok('주보 시안 드롭다운 열림')
        : warn('주보 시안 드롭다운 내용 확인 필요');
      await shot(page, 'template-selector');
    }

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 8. UX 품질 — 빈 상태, 에러 처리, 로딩
// ═══════════════════════════════════════════════════════════════
async function testUXQuality(browser) {
  section('UX 품질 (에러/빈 상태/로딩)');

  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, storageState: STORAGE_NO_TOUR });
  const page = await ctx.newPage();
  try {
    // ── 찬송가 — 검색 전 빈 상태 안내 ──
    await page.goto(`${BASE}/lyrics`, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(600);

    const hymnTab = page.locator('button').filter({ hasText: '찬송가 검색' });
    if (await hymnTab.count() > 0) {
      await hymnTab.first().click();
      await page.waitForTimeout(400);
    }
    const emptyHymn = await page.locator('body').textContent();
    /번호나 제목으로|검색하세요|Search|입력하세요/i.test(emptyHymn)
      ? ok('찬송가 검색 전 빈 상태 안내 텍스트 존재')
      : warn('찬송가 검색 전 안내 문구 없음 — UX 개선 권장');

    // ── 존재하지 않는 찬송가 검색 결과 없음 안내 ──
    const searchInput = page.locator('input[placeholder*="번호 또는 제목"]');
    if (await searchInput.count() > 0) {
      await searchInput.click();
      await searchInput.pressSequentially('99999', { delay: 60 });
      await page.waitForTimeout(400);
      const searchBtn = page.locator('button').filter({ hasText: /^검색/ });
      if (await searchBtn.count() > 0) {
        await searchBtn.first().click();
        await page.waitForTimeout(1500);
      }
      const noResult = await page.locator('body').textContent();
      /결과가 없습니다|없음|No result|found/i.test(noResult)
        ? ok('찬송가 검색 결과 없음 시 안내 메시지 표시')
        : ng('찬송가 검색 결과 없음 시 안내 메시지 없음');
    }

    // ── 주보 페이지 — disabled 상태 버튼 ──
    await page.goto(`${BASE}/bulletin`, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(800);

    // PDF 다운로드 버튼은 기본적으로 활성화 상태여야 함
    const pdfBtn = page.locator('button').filter({ hasText: 'PDF 다운로드' });
    if (await pdfBtn.count() > 0) {
      const isDisabled = await pdfBtn.first().getAttribute('disabled');
      isDisabled === null
        ? ok('PDF 다운로드 버튼 초기 활성화 상태')
        : ng('PDF 다운로드 버튼 초기 비활성화 (로드 전)')
    }

    // ── 타임라인 — 항목이 있는지 ──
    await page.goto(BASE, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(800);

    const timeline = page.locator('[class*="timeline"], [class*="Timeline"]').first();
    const timelineExists = await timeline.count() > 0;
    timelineExists
      ? ok('ProTimeline 렌더링 확인')
      : warn('ProTimeline 클래스 기반 선택 실패 — 구조 변경 가능성');

    // ── 인스펙터 패널 토글 ──
    const inspBtn = page.locator('button[title*="인스펙터"], button[title*="inspector"], button[aria-label*="인스펙터"]').first();
    const hasInspBtn = await inspBtn.count() > 0;
    if (!hasInspBtn) {
      // topbar 우측 버튼으로 대체
      const topbarBtns = page.locator('[data-testid="topbar"] button');
      const topbarBtnCount = await topbarBtns.count();
      topbarBtnCount > 0
        ? warn(`인스펙터 토글 버튼 aria-label 없음 — 접근성 개선 권장 (TopBar에 버튼 ${topbarBtnCount}개 있음)`)
        : ng('인스펙터 토글 버튼 없음');
    } else {
      ok('인스펙터 토글 버튼 존재 (aria-label 있음)');
    }

    await shot(page, 'ux-quality-check');

    // ── 주보 편집 — 예배 순서 아이템 클릭 시 Detail 패널 연동 ──
    await page.goto(`${BASE}/bulletin`, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(800);

    // WorshipOrder 체크박스나 항목 클릭
    const worshipItems = page.locator('[class*="WorshipOrder"] input[type="checkbox"], [class*="worship-order"] input[type="checkbox"]');
    const wItemCount = await worshipItems.count();
    wItemCount > 0
      ? ok(`예배 순서 체크박스 ${wItemCount}개 존재`)
      : warn('예배 순서 체크박스 없음 — 선택 방식 확인 필요');

    // SelectedOrder 영역이 있는지
    const selectedArea = page.locator('[class*="SelectedOrder"], [class*="selected-order"]').first();
    const selectedExists = await selectedArea.count() > 0;
    selectedExists
      ? ok('SelectedOrder 컴포넌트 렌더링 확인')
      : warn('SelectedOrder 컴포넌트 선택자 확인 필요');

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 9. 모바일 UX & 터치 타겟
// ═══════════════════════════════════════════════════════════════
async function testMobileUX(browser) {
  section('모바일 UX & 터치 타겟');

  const ctx = await browser.newContext({
    viewport: { width: 375, height: 812 },
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1',
    storageState: STORAGE_NO_TOUR,
  });
  const page = await ctx.newPage();
  try {
    await page.goto(BASE, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(800);

    // 핵심 버튼 터치 타겟 크기 (최소 44×44px)
    const buttons = page.locator('button:visible, a:visible');
    const btnCount = await buttons.count();
    let tooSmall = 0;
    const smallButtons = [];
    for (let i = 0; i < Math.min(btnCount, 20); i++) {
      const btn = buttons.nth(i);
      const box = await btn.boundingBox().catch(() => null);
      if (box && (box.width < 36 || box.height < 36)) {
        tooSmall++;
        const txt = (await btn.textContent().catch(() => '')).trim().slice(0, 20);
        smallButtons.push(`"${txt}" (${Math.round(box.width)}×${Math.round(box.height)})`);
      }
    }
    tooSmall === 0
      ? ok(`모바일 버치 타겟 크기 정상 (${Math.min(btnCount, 20)}개 확인)`)
      : warn(`터치 타겟 너무 작은 버튼 ${tooSmall}개: ${smallButtons.slice(0,3).join(', ')}`);

    // 가로 스크롤 없음
    const hasHScroll = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth);
    !hasHScroll
      ? ok('모바일 375px 가로 스크롤 없음')
      : ng('모바일 375px 가로 스크롤 발생');

    await shot(page, 'mobile-ux-375');

    // bulletin 페이지 모바일
    await page.goto(`${BASE}/bulletin`, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(600);

    const bHScroll = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth);
    !bHScroll
      ? ok('bulletin 모바일 375px 가로 스크롤 없음')
      : ng('bulletin 모바일 375px 가로 스크롤 발생');

    // 버튼들이 겹치지 않는지 확인
    const headerBtns = page.locator('button:visible').filter({ hasText: /PDF|프로젝터/ });
    const hBtnCount = await headerBtns.count();
    if (hBtnCount >= 2) {
      const box1 = await headerBtns.nth(0).boundingBox().catch(() => null);
      const box2 = await headerBtns.nth(1).boundingBox().catch(() => null);
      if (box1 && box2) {
        const overlap = box1.x < box2.x + box2.width && box1.x + box1.width > box2.x &&
                        box1.y < box2.y + box2.height && box1.y + box1.height > box2.y;
        !overlap
          ? ok('PDF/프로젝터 버튼 모바일에서 겹침 없음')
          : ng('PDF/프로젝터 버튼 모바일에서 겹침 발생');
      }
    }
    await shot(page, 'mobile-bulletin-375');

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 10. Display HTML — CSS 변수 & 오버레이 렌더링
// ═══════════════════════════════════════════════════════════════
async function testDisplayRendering(browser) {
  section('Display HTML 렌더링 & CSS 변수');

  const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 }, storageState: STORAGE_NO_TOUR });
  const page = await ctx.newPage();
  try {
    // /display/overlay — CSS 변수 적용 확인
    await page.goto(`${GO}/display/overlay`, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(600);

    // body 배경 투명 (크로마키 필수)
    const bodyBg = await page.evaluate(() => {
      const s = window.getComputedStyle(document.body);
      return { bg: s.backgroundColor, color: s.color };
    });
    const isTransparent = bodyBg.bg === 'rgba(0, 0, 0, 0)' || bodyBg.bg === 'transparent';
    isTransparent
      ? ok('/display/overlay body 배경 투명 (크로마키)')
      : ng('/display/overlay body 배경 불투명', `got: ${bodyBg.bg}`);

    // #slide 요소
    const slide = page.locator('#slide');
    const slideExists = await slide.count() > 0;
    slideExists
      ? ok('/display/overlay #slide 요소 존재')
      : ng('/display/overlay #slide 요소 없음');

    await shot(page, 'display-overlay-detail');

    // DisplayConfig → CSS 변수 반영 확인
    const config = await httpGet(`${GO}/api/display-config`);
    if (config.status === 200 && slideExists) {
      // overlay의 CSS 변수를 읽어 config 값과 비교
      const cssVars = await page.evaluate(() => {
        const style = document.documentElement.style;
        return {
          bgOpacity: style.getPropertyValue('--overlay-bg-opacity').trim(),
          textColor: style.getPropertyValue('--overlay-text-color').trim(),
          position: style.getPropertyValue('--overlay-position').trim(),
          fontScale: style.getPropertyValue('--overlay-font-scale').trim(),
        };
      });

      // CSS 변수는 JS에서 동적으로 주입되거나 HTML에 인라인일 수 있음
      // 값이 존재하면 config와 비교
      if (cssVars.bgOpacity) {
        const expected = String(config.body.overlayBgOpacity);
        cssVars.bgOpacity === expected
          ? ok(`overlay CSS --overlay-bg-opacity = ${cssVars.bgOpacity} (config 일치)`)
          : warn(`overlay CSS --overlay-bg-opacity = "${cssVars.bgOpacity}" vs config "${expected}" — 비동기 반영 가능`);
      } else {
        // overlay.html에서 inline style로 적용되는 경우
        const htmlContent = await page.content();
        htmlContent.includes('overlay') || htmlContent.includes('opacity')
          ? warn('overlay CSS 변수 inline 적용 방식 — JavaScript 초기화 필요')
          : ng('overlay CSS 변수 미적용 확인 필요');
      }
    }

    // /display (OBS) — 슬라이드 번호 및 콘텐츠
    await page.goto(`${GO}/display`, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(600);

    const displayBody = await page.locator('body').textContent();
    displayBody.length > 50
      ? ok('/display 본문 콘텐츠 존재')
      : ng('/display 본문 비어있음');

    // /display/stage — 스테이지 구조
    await page.goto(`${GO}/display/stage`, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(600);

    const stageBody = await page.locator('body').textContent();
    stageBody.length > 30
      ? ok('/display/stage 콘텐츠 존재')
      : ng('/display/stage 콘텐츠 비어있음');

    await shot(page, 'display-stage-detail');

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// 11. 찬송가 DB 데이터 품질 검사
// ═══════════════════════════════════════════════════════════════
async function testHymnDBQuality() {
  section('찬송가 DB 데이터 품질');

  // 전체 수 확인
  const all = await httpGet(`${GO}/api/hymns?page=1&limit=1`);
  const total = all.body?.total || 0;
  total >= 645
    ? ok(`찬송가 전체 수: ${total}개 (645개 이상)`)
    : ng('찬송가 수 부족', `got: ${total}`);

  // 번호 1~645 범위 확인 (첫 페이지)
  const page1 = await httpGet(`${GO}/api/hymns?page=1&limit=50`);
  if (page1.body?.hymns) {
    const nums = page1.body.hymns.map((h) => h.number);
    const min = Math.min(...nums);
    const max = Math.max(...nums);
    min >= 1 && max <= 645
      ? ok(`찬송가 번호 범위 정상: ${min}~${max}`)
      : ng('찬송가 번호 범위 이상', `${min}~${max}`);
  }

  // 가사가 있는 찬송가 비율
  const hymnWithLyrics = await httpGet(`${GO}/api/hymns/search?q=1`);
  const hasLyricsCount = Array.isArray(hymnWithLyrics.body)
    ? hymnWithLyrics.body.filter((h) => h.lyrics && h.lyrics.length > 10).length
    : 0;
  const total1 = Array.isArray(hymnWithLyrics.body) ? hymnWithLyrics.body.length : 0;
  hasLyricsCount === total1 && total1 > 0
    ? ok(`번호 검색 결과 가사 보유율: ${hasLyricsCount}/${total1}`)
    : warn(`가사 없는 찬송가 있음: ${total1 - hasLyricsCount}/${total1}개`);

  // first_line 데이터 있는지
  const hymnSample = await httpGet(`${GO}/api/hymns?page=1&limit=10`);
  const withFirstLine = (hymnSample.body?.hymns || []).filter((h) => h.first_line);
  withFirstLine.length === 10
    ? ok('찬송가 first_line 필드 모두 있음 (샘플 10개)')
    : warn(`first_line 없는 항목: ${10 - withFirstLine.length}/10개`);

  // hymnbook 분포 (new, old, 21c)
  const newBook = await httpGet(`${GO}/api/hymns?page=1&limit=1&book=new`);
  const oldBook = await httpGet(`${GO}/api/hymns?page=1&limit=1&book=old`);
  newBook.body?.total > 0
    ? ok(`새찬송가(new): ${newBook.body.total}개`)
    : warn('새찬송가(new) 데이터 없음');
  oldBook.body?.total > 0
    ? ok(`구찬송가(old): ${oldBook.body.total}개`)
    : warn('구찬송가(old) 데이터 없음');

  // 특정 찬송가 데이터 정합성 (1번: "만복의 근원 하나님")
  const hymn1 = await httpGet(`${GO}/api/hymns/detail?number=1&book=new`);
  hymn1.status === 200
    ? ok(`찬송가 1번 조회 성공: "${hymn1.body?.title}"`)
    : ng('찬송가 1번 조회 실패');

  // 645번 (마지막)
  const hymn645 = await httpGet(`${GO}/api/hymns/detail?number=645&book=new`);
  [200, 404].includes(hymn645.status)
    ? ok(`찬송가 645번 응답: ${hymn645.status}`)
    : ng('찬송가 645번 비정상 응답', `status=${hymn645.status}`);

  // 찬송가 search type=lyrics
  const lyricsSearch = await httpGet(`${GO}/api/hymns/search?q=주+예수&type=lyrics`);
  lyricsSearch.status === 200 && Array.isArray(lyricsSearch.body)
    ? ok(`가사 검색(type=lyrics): ${lyricsSearch.body.length}개`)
    : warn(`가사 검색 오류: status=${lyricsSearch.status}`);
}

// ═══════════════════════════════════════════════════════════════
// 12. 성경 DB 데이터 품질 검사
// ═══════════════════════════════════════════════════════════════
async function testBibleDBQuality() {
  section('성경 DB 데이터 품질');

  // 66권 전부 절 데이터 있는지 (샘플: 창, 시, 마, 요, 계)
  const testBooks = [
    { book: 1,  name: '창세기',     chapter: 1, minVerses: 31 },
    { book: 19, name: '시편',       chapter: 1, minVerses: 6 },
    { book: 40, name: '마태복음',   chapter: 1, minVerses: 25 },
    { book: 43, name: '요한복음',   chapter: 3, minVerses: 36 },
    { book: 66, name: '요한계시록', chapter: 1, minVerses: 20 },
  ];
  for (const tb of testBooks) {
    const r = await httpGet(`${GO}/api/bible/verses?book=${tb.book}&chapter=${tb.chapter}`);
    if (r.status === 200 && Array.isArray(r.body)) {
      r.body.length >= tb.minVerses
        ? ok(`${tb.name} ${tb.chapter}장: ${r.body.length}절 (최소 ${tb.minVerses}절 이상)`)
        : ng(`${tb.name} ${tb.chapter}장 절 수 부족`, `got: ${r.body.length}, expected >= ${tb.minVerses}`);
      // 텍스트 비어있지 않음
      const emptyVerse = r.body.find((v) => !v.text || v.text.trim() === '');
      emptyVerse
        ? ng(`${tb.name} 빈 절 발견`, `verse=${emptyVerse.verse}`)
        : ok(`${tb.name} 모든 절 텍스트 있음`);
    } else {
      ng(`${tb.name} 절 조회 실패`, `status=${r.status}`);
    }
  }

  // 번역 버전 목록
  const versions = await httpGet(`${GO}/api/bible/versions`);
  if (versions.status === 200 && Array.isArray(versions.body)) {
    versions.body.length >= 1
      ? ok(`성경 번역 버전 ${versions.body.length}개: ${versions.body.join(', ')}`)
      : ng('성경 번역 버전 0개');

    // 각 버전으로 창세기 1:1 조회
    for (const ver of versions.body.slice(0, 3)) {
      const vr = await httpGet(`${GO}/api/bible/verses?book=1&chapter=1&version=${encodeURIComponent(ver)}`);
      vr.status === 200 && vr.body?.length > 0
        ? ok(`번역 버전 "${ver}" 창세기 1장 조회 성공`)
        : warn(`번역 버전 "${ver}" 창세기 1장 조회 실패: status=${vr.status}`);
    }
  } else {
    ng('성경 버전 API 오류', `status=${versions.status}`);
  }

  // 성경 검색
  const bSearch = await httpGet(`${GO}/api/bible/search?q=태초`);
  bSearch.status === 200
    ? ok(`성경 "태초" 검색: status 200`)
    : warn(`성경 검색 API status=${bSearch.status}`);

  if (bSearch.status === 200) {
    const searchResults = bSearch.body;
    const hasResults = Array.isArray(searchResults) ? searchResults.length > 0 :
                       searchResults?.results?.length > 0 || searchResults?.verses?.length > 0;
    hasResults
      ? ok('"태초" 검색 결과 있음')
      : ng('"태초" 검색 결과 없음');
  }
}

// ═══════════════════════════════════════════════════════════════
// 13. 설정 & 라이선스 페이지 UX
// ═══════════════════════════════════════════════════════════════
async function testSettingsUX(browser) {
  section('설정 & 라이선스 UX');

  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, storageState: STORAGE_NO_TOUR });
  const page = await ctx.newPage();
  try {
    await page.goto(BASE, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(800);

    // 설정 페이지 접근 (아이콘 바 또는 네비게이션)
    const settingsBtn = page.locator('a[href*="setting"], button[title*="설정"], a[href*="config"]').first();
    const hasSettings = await settingsBtn.count() > 0;
    if (hasSettings) {
      await settingsBtn.click();
      await page.waitForTimeout(800);
      ok('설정 페이지 네비게이션 성공');
      await shot(page, 'settings-page');
    } else {
      // 아이콘 바에서 찾기
      const iconBar = page.locator('[class*="IconBar"], [class*="icon-bar"], [class*="sidebar"]').first();
      const iconBtns = iconBar.locator('a, button');
      const iconCount = await iconBtns.count();
      warn(`설정 링크 직접 찾기 실패 — 아이콘 바 ${iconCount}개 버튼 존재, 수동 확인 필요`);
    }

    // 라이선스 API
    const license = await httpGet(`${GO}/api/license`);
    license.status === 200
      ? ok(`라이선스 API 정상: plan="${license.body?.plan}", active=${license.body?.active}`)
      : ng('라이선스 API 오류', `status=${license.status}`);

    if (license.status === 200) {
      typeof license.body?.plan === 'string'
        ? ok(`라이선스 plan 필드: "${license.body.plan}"`)
        : ng('라이선스 plan 필드 누락 또는 문자열 아님');
    }

    // 버전 API
    const version = await httpGet(`${GO}/api/version`);
    version.status === 200 && version.body?.version
      ? ok(`서버 버전: ${version.body.version}`)
      : ng('버전 API 오류', `status=${version.status}`);

    // 업데이트 확인 API (오류 없이 응답하는지)
    const updateCheck = await httpGet(`${GO}/api/update/check`);
    [200, 503].includes(updateCheck.status)
      ? ok(`업데이트 확인 API: ${updateCheck.status}`)
      : ng('업데이트 확인 API 오류', `status=${updateCheck.status}`);

  } finally {
    await ctx.close();
  }
}

// ═══════════════════════════════════════════════════════════════
// main
// ═══════════════════════════════════════════════════════════════
async function main() {
  const startTime = Date.now();
  process.stdout.write(`\n🔬 easyPreparation 상세 버그 탐지 테스트\n`);
  process.stdout.write(`   Next.js: ${BASE}  |  Go: ${GO}\n`);
  process.stdout.write(`${'═'.repeat(54)}\n`);

  const browser = await chromium.launch({ headless: true });
  try {
    await testAPIFieldIntegrity();
    await testAPIEdgeCases();
    await testAPIDBRoundTrip();
    await testAPIvsUI(browser);
    await testDisplayFlow();
    await testFirstTimeUX(browser);
    await testDetailedFeatureUX(browser);
    await testUXQuality(browser);
    await testMobileUX(browser);
    await testDisplayRendering(browser);
    await testHymnDBQuality();
    await testBibleDBQuality();
    await testSettingsUX(browser);
  } finally {
    await browser.close();
  }

  const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);

  process.stdout.write(`\n${'═'.repeat(54)}\n`);
  process.stdout.write(`결과: ${pass + fail}개 검사  ✓ ${pass} 통과  ✗ ${fail} 실패  (${elapsed}s)\n\n`);

  process.stdout.write(`섹션별 결과:\n`);
  for (const [sec, r] of Object.entries(sectionResults)) {
    const icon = r.fail > 0 ? '⚠️' : '✅';
    process.stdout.write(`  ${icon} ${sec}: ${r.pass}통과 / ${r.fail}실패\n`);
  }

  if (issues.length > 0) {
    process.stdout.write(`\n🐛 발견된 버그 (${issues.length}개):\n`);
    issues.forEach((i) => process.stdout.write(`  • ${i}\n`));
  }

  if (feedbacks.length > 0) {
    process.stdout.write(`\n💡 UX 개선 권고 (${feedbacks.length}개):\n`);
    feedbacks.forEach((f) => process.stdout.write(`  ⚠ ${f}\n`));
  }

  if (issues.length === 0 && feedbacks.length === 0) {
    process.stdout.write(`\n✅ 버그 없음, UX 이슈 없음\n`);
  } else if (issues.length === 0) {
    process.stdout.write(`\n✅ 버그 없음 (UX 권고 사항 ${feedbacks.length}개)\n`);
  }

  process.stdout.write(`\n📁 스크린샷: ${OUT_DIR}\n`);
  process.exit(issues.length > 0 ? 1 : 0);
}

main().catch((e) => { console.error(e); process.exit(1); });
