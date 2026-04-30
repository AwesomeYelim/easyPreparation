#!/usr/bin/env node
// UI 기능 테스트 — Playwright 병렬 실행
// 실행: cd ui && npm run test:ui
// 전제: make dev 로 서버 실행 중 (Go :8080 + Next.js :3000)

const { chromium } = require('@playwright/test');

const TIMEOUT = 10000;
const results = [];

function pass(name, detail = '') {
  results.push({ name, ok: true, detail });
  console.log(`  ✓ ${name}${detail ? ': ' + detail : ''}`);
}

function fail(name, detail = '') {
  results.push({ name, ok: false, detail });
  console.error(`  ✗ ${name}${detail ? ': ' + detail : ''}`);
}

// ── 테스트 정의 ────────────────────────────────────────────────
async function testMain(browser) {
  console.log('\n[main] 주예배 페이지');
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await ctx.newPage();
  try {
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle', timeout: TIMEOUT });

    // 1. 예배 순서 패널 렌더링
    const seqPanel = await page.$('[class*="ProSequencePanel"], [class*="sequence"], [data-testid="seq-panel"]');
    seqPanel ? pass('예배 순서 패널 렌더링') : fail('예배 순서 패널 없음');

    // 2. 탭 네비게이션 (주예배/애찬/수요)
    const tabs = await page.$$('[class*="tab"], [role="tab"], button[class*="Tab"]');
    tabs.length >= 2 ? pass('탭 네비게이션', `${tabs.length}개`) : fail('탭 부족', `${tabs.length}개`);

    // 3. API 응답 — worship order 로드
    const apiOk = await page.evaluate(async () => {
      try {
        const r = await fetch('http://localhost:8080/api/worship-order');
        return r.ok;
      } catch { return false; }
    });
    apiOk ? pass('API /worship-order 응답') : fail('API /worship-order 실패');

    // 4. 스크롤 — 가로 스크롤 없음
    const hasHScroll = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 5);
    hasHScroll ? fail('가로 스크롤 발생 (overflow)') : pass('가로 스크롤 없음');

  } finally {
    await ctx.close();
  }
}

async function testMobile(browser) {
  console.log('\n[mobile] 모바일 뷰');
  const ctx = await browser.newContext({ viewport: { width: 375, height: 812 } });
  const page = await ctx.newPage();
  try {
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle', timeout: TIMEOUT });

    // 5. 모바일 가로 스크롤 없음
    const hasHScroll = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 5);
    hasHScroll ? fail('모바일 가로 스크롤 발생') : pass('모바일 가로 스크롤 없음');

    // 6. 모바일에서 주요 요소 visible
    const mainEl = await page.$('main, [class*="Shell"], [class*="Layout"]');
    mainEl ? pass('모바일 레이아웃 렌더링') : fail('모바일 레이아웃 없음');

  } finally {
    await ctx.close();
  }
}

async function testDisplay(browser) {
  console.log('\n[display] OBS 디스플레이');
  const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await ctx.newPage();
  try {
    const res = await page.goto('http://localhost:8080/display', { waitUntil: 'networkidle', timeout: TIMEOUT });

    // 7. 200 응답
    res && res.status() === 200 ? pass('display 200 응답') : fail('display 응답 오류', String(res?.status()));

    // 8. 빈 화면 아님
    const bodyText = await page.evaluate(() => document.body.innerText.trim());
    bodyText.length > 10 ? pass('display 컨텐츠 존재') : fail('display 빈 화면');

    // 9. WebSocket 연결 (display WS)
    const wsOk = await page.evaluate(() => {
      return new Promise(resolve => {
        const ws = new WebSocket('ws://localhost:8080/ws/display');
        ws.onopen = () => { ws.close(); resolve(true); };
        ws.onerror = () => resolve(false);
        setTimeout(() => resolve(false), 3000);
      });
    });
    wsOk ? pass('Display WebSocket 연결') : fail('Display WebSocket 실패');

  } catch (e) {
    fail('display 페이지 접근 실패', e.message);
  } finally {
    await ctx.close();
  }
}

async function testBible(browser) {
  console.log('\n[bible] 성경 페이지');
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await ctx.newPage();
  try {
    await page.goto('http://localhost:3000/bible', { waitUntil: 'networkidle', timeout: TIMEOUT });

    // 10. 성경 선택 UI 렌더링
    const bibleEl = await page.$('select, [class*="bible"], [class*="Bible"]');
    bibleEl ? pass('성경 선택 UI 렌더링') : fail('성경 선택 UI 없음');

  } finally {
    await ctx.close();
  }
}

// ── 실행 ───────────────────────────────────────────────────────
async function main() {
  const browser = await chromium.launch({ headless: true });

  await Promise.all([
    testMain(browser),
    testMobile(browser),
    testDisplay(browser),
    testBible(browser),
  ]);

  await browser.close();

  const failed = results.filter(r => !r.ok);
  console.log(`\n${'─'.repeat(40)}`);
  console.log(`결과: ${results.length - failed.length}/${results.length} 통과`);
  if (failed.length) {
    console.error('실패 항목:');
    failed.forEach(f => console.error(`  - ${f.name}: ${f.detail}`));
    process.exit(1);
  }
  console.log('✅ 모든 기능 테스트 통과');
}

main().catch(e => { console.error(e.message); process.exit(1); });
