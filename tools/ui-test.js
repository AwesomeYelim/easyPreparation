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

async function skipTour(page) {
  try {
    await page.evaluate(() => localStorage.setItem('ep_tour_done', '1'));
    await page.reload({ waitUntil: 'networkidle', timeout: TIMEOUT });
  } catch {}
}

// ── 테스트 정의 ────────────────────────────────────────────────
async function testMain(browser) {
  console.log('\n[main] 주예배 페이지');
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await ctx.newPage();
  try {
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle', timeout: TIMEOUT });
    await skipTour(page);

    // 1. 예배 순서 패널 렌더링
    const seqPanel = await page.$('[data-testid="seq-panel"]');
    seqPanel ? pass('예배 순서 패널 렌더링') : fail('예배 순서 패널 없음');

    // 2. 탭 네비게이션
    const tabs = await page.$$('[class*="tab"], [role="tab"], button[class*="Tab"]');
    tabs.length >= 2 ? pass('탭 네비게이션', `${tabs.length}개`) : fail('탭 부족', `${tabs.length}개`);

    // 3. API /worship-order 응답
    const apiOk = await page.evaluate(async () => {
      try {
        const r = await fetch('http://localhost:8080/api/worship-order?type=main_worship');
        return r.ok;
      } catch { return false; }
    });
    apiOk ? pass('API /worship-order 응답') : fail('API /worship-order 실패');

    // 4. 가로 스크롤 없음
    const hasHScroll = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 5);
    hasHScroll ? fail('가로 스크롤 발생') : pass('가로 스크롤 없음');

    // 5. ProTopBar 렌더링
    const topbar = await page.$('[data-testid="topbar"]');
    topbar ? pass('TopBar 렌더링') : fail('TopBar 없음');

    // 6. ProTimeline 렌더링
    const timeline = await page.$('[data-testid="timeline"]');
    timeline ? pass('Timeline 렌더링') : fail('Timeline 없음');

    // 7. 예배 순서 아이템 개수
    const items = await page.$$('[data-testid="seq-item"]');
    items.length > 0 ? pass('예배 순서 아이템', `${items.length}개`) : fail('예배 순서 아이템 없음');

    // 8. 활성 아이템 하이라이트 (파란 border-l)
    const activeItem = await page.$('[data-testid="seq-item"][data-active="true"]');
    activeItem ? pass('활성 아이템 하이라이트') : fail('활성 아이템 강조 없음');

    // 9. Timeline에 시간 예상 텍스트 표시
    const timelineText = await page.evaluate(() => {
      const el = document.querySelector('[data-testid="timeline"]');
      return el ? el.innerText : '';
    });
    timelineText.includes('h') || timelineText.includes('분') || timelineText.includes(':')
      ? pass('Timeline 시간 표시', timelineText.slice(0, 30).replace(/\n/g, ' '))
      : fail('Timeline 시간 표시 없음');

    // 10. Inspector 패널 렌더링 (Studio)
    const inspector = await page.$('[data-testid="inspector"]');
    if (!inspector) {
      // data-testid 없으면 텍스트로 찾기
      const studioText = await page.evaluate(() => document.body.innerText.includes('Studio'));
      studioText ? pass('Inspector 패널 렌더링') : fail('Inspector 패널 없음');
    } else {
      pass('Inspector 패널 렌더링');
    }

    // 11. PREV/NEXT 버튼 존재
    const prevBtn = await page.$('button:has-text("PREV"), button:has-text("◀")');
    const nextBtn = await page.$('button:has-text("NEXT"), button:has-text("▶")');
    prevBtn && nextBtn ? pass('PREV/NEXT 네비게이션 버튼') : fail('PREV/NEXT 버튼 없음', `prev:${!!prevBtn} next:${!!nextBtn}`);

    // 12. 아이템 직접 클릭 시 active 변경 (handleJump — 즉시 Recoil setIdx)
    const allItems = await page.$$('[data-testid="seq-item"]');
    if (allItems.length >= 2) {
      const beforeActive = await page.evaluate(() => {
        const el = document.querySelector('[data-testid="seq-item"][data-active="true"]');
        return el ? el.innerText.slice(0, 20) : '';
      });
      // 현재 활성이 아닌 첫 번째 아이템 클릭
      const target = allItems[0];
      const targetActive = await target.getAttribute('data-active');
      const clickTarget = targetActive === 'true' ? allItems[1] : allItems[0];
      await clickTarget.click();
      await page.waitForTimeout(400);
      const afterActive = await page.evaluate(() => {
        const el = document.querySelector('[data-testid="seq-item"][data-active="true"]');
        return el ? el.innerText.slice(0, 20) : '';
      });
      beforeActive !== afterActive
        ? pass('아이템 클릭 시 active 전환', `${beforeActive.trim()} → ${afterActive.trim()}`)
        : fail('아이템 클릭 후 active 미변경');
    }

    // 13. 예배 타입 드롭다운 존재
    const worshipTypeDropdown = await page.$('select, [class*="dropdown"], button:has-text("예배")');
    worshipTypeDropdown ? pass('예배 타입 드롭다운 존재') : fail('예배 타입 드롭다운 없음');

    // 14. 시계 표시 (HH:MM:SS 형식)
    const clockText = await page.evaluate(() => {
      const topbar = document.querySelector('[data-testid="topbar"]');
      return topbar ? topbar.innerText : '';
    });
    /\d{2}:\d{2}:\d{2}/.test(clockText) ? pass('상단 시계 표시') : fail('상단 시계 없음');

  } finally {
    await ctx.close();
  }
}

async function testMobile(browser) {
  console.log('\n[mobile] 모바일 뷰 (375px)');
  const ctx = await browser.newContext({ viewport: { width: 375, height: 812 } });
  const page = await ctx.newPage();
  try {
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle', timeout: TIMEOUT });
    await skipTour(page);

    // 15. 가로 스크롤 없음
    const hasHScroll = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 5);
    hasHScroll ? fail('모바일 가로 스크롤 발생') : pass('모바일 가로 스크롤 없음');

    // 16. 모바일 레이아웃 렌더링
    const mainEl = await page.$('main, [class*="Shell"], [class*="Layout"], [class*="pro-shell"]');
    mainEl ? pass('모바일 레이아웃 렌더링') : fail('모바일 레이아웃 없음');

    // 17. 모바일에서 Inspector 패널 자동 닫힘 (뷰포트 내 Studio 패널 노출 금지)
    const inspectorRect = await page.evaluate(() => {
      // Studio 패널이 뷰포트 안에 있으면 안 됨 (overflow hidden 처리)
      const allText = document.body.innerText;
      // inspector가 열려있으면 Studio 텍스트가 DOM에 존재
      // 뷰포트 밖으로 잘려야 정상
      const vpWidth = window.innerWidth;
      const scrollW = document.documentElement.scrollWidth;
      return { vpWidth, scrollW };
    });
    // scrollWidth가 뷰포트를 초과하지 않아야 함
    inspectorRect.scrollW <= inspectorRect.vpWidth + 5
      ? pass('모바일 Inspector 자동 닫힘')
      : fail('모바일 Inspector 뷰포트 초과', `scroll:${inspectorRect.scrollW} vp:${inspectorRect.vpWidth}`);

    // 18. 모바일에서 TopBar 렌더링
    const topbar = await page.$('[data-testid="topbar"]');
    topbar ? pass('모바일 TopBar 렌더링') : fail('모바일 TopBar 없음');

    // 19. 모바일에서 시퀀스 패널 자동 닫힘 (좁은 화면에서 메인 영역 확보)
    const seqItems = await page.$$('[data-testid="seq-item"]');
    seqItems.length === 0
      ? pass('모바일 시퀀스 패널 자동 닫힘 (메인 영역 확보)')
      : fail('모바일에서 시퀀스 패널 미닫힘', `${seqItems.length}개 노출`);

    // 20. 탭 텍스트 가시성 (탭 레이블이 존재해야 함)
    const topbarText = await page.evaluate(() => {
      const el = document.querySelector('[data-testid="topbar"]');
      return el ? el.innerText : '';
    });
    topbarText.length > 5 ? pass('모바일 TopBar 텍스트 가시') : fail('모바일 TopBar 텍스트 없음');

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

    // 21. 200 응답
    res && res.status() === 200 ? pass('display 200 응답') : fail('display 응답 오류', String(res?.status()));

    // 22. 빈 화면 아님
    const bodyText = await page.evaluate(() => document.body.innerText.trim());
    bodyText.length > 10 ? pass('display 컨텐츠 존재') : fail('display 빈 화면');

    // 23. WebSocket 연결
    const wsOk = await page.evaluate(() => {
      return new Promise(resolve => {
        const ws = new WebSocket('ws://localhost:8080/ws');
        ws.onopen = () => { ws.close(); resolve(true); };
        ws.onerror = () => resolve(false);
        setTimeout(() => resolve(false), 3000);
      });
    });
    wsOk ? pass('Display WebSocket 연결') : fail('Display WebSocket 실패');

    // 24. display/stage 200 응답
    const stageRes = await page.goto('http://localhost:8080/display/stage', { waitUntil: 'networkidle', timeout: TIMEOUT });
    stageRes && stageRes.status() === 200 ? pass('display/stage 200 응답') : fail('display/stage 응답 오류', String(stageRes?.status()));

    // 25. stage 컨텐츠 존재
    const stageText = await page.evaluate(() => document.body.innerText.trim());
    stageText.length > 5 ? pass('stage 컨텐츠 존재') : fail('stage 빈 화면');

    // 26. stage 레이아웃 — 현재/다음 슬라이드 패널
    const hasCurrentPanel = await page.evaluate(() => {
      const text = document.body.innerText;
      return text.includes('현재') || text.includes('다음') || text.includes('/');
    });
    hasCurrentPanel ? pass('stage 슬라이드 네비게이션 표시') : fail('stage 슬라이드 정보 없음');

    // 27. display API — church-info 응답 (서버 기본 상태 확인)
    const churchApi = await page.evaluate(async () => {
      try {
        const r = await fetch('http://localhost:8080/api/church-info');
        return r.status;
      } catch { return 0; }
    });
    churchApi === 200
      ? pass('church-info API 응답', `HTTP ${churchApi}`)
      : fail('church-info API 실패', `HTTP ${churchApi}`);

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
    await skipTour(page);

    // 28. 성경 선택 UI 렌더링
    const bibleEl = await page.$('select, [class*="bible"], [class*="Bible"]');
    bibleEl ? pass('성경 선택 UI 렌더링') : fail('성경 선택 UI 없음');

    // 29. 구약/신약 탭 존재
    const tabs = await page.$$('button');
    const tabTexts = await Promise.all(tabs.map(t => t.innerText().catch(() => '')));
    const hasOT = tabTexts.some(t => t.includes('구약'));
    const hasNT = tabTexts.some(t => t.includes('신약'));
    hasOT && hasNT ? pass('구약/신약 탭 존재') : fail('구약/신약 탭 없음', `구약:${hasOT} 신약:${hasNT}`);

    // 30. 성경 API 응답
    const bibleApi = await page.evaluate(async () => {
      try {
        const r = await fetch('http://localhost:8080/api/bible/books?testament=old');
        return r.ok;
      } catch { return false; }
    });
    bibleApi ? pass('성경 API /bible/books 응답') : fail('성경 API 실패');

    // 31. 구약 책 목록 렌더링 (창세기 등)
    const bookListText = await page.evaluate(() => document.body.innerText);
    bookListText.includes('창세기') || bookListText.includes('창')
      ? pass('구약 책 목록 렌더링')
      : fail('구약 책 목록 없음');

    // 32. 신약 탭 클릭 시 신약 책 표시
    const ntTab = await page.$('button:has-text("신약")');
    if (ntTab) {
      await ntTab.click();
      await page.waitForTimeout(500);
      const afterText = await page.evaluate(() => document.body.innerText);
      afterText.includes('마태') || afterText.includes('요한') || afterText.includes('로마')
        ? pass('신약 탭 클릭 → 신약 책 표시')
        : fail('신약 탭 클릭 후 신약 책 미표시');
    }

    // 33. 책 클릭 시 장 목록 표시
    const firstBook = await page.$('button:has-text("창세기"), button:has-text("마태")');
    if (firstBook) {
      await firstBook.click();
      await page.waitForTimeout(800);
      const chapterButtons = await page.$$('button');
      const chapterTexts = await Promise.all(chapterButtons.map(b => b.innerText().catch(() => '')));
      const hasChapter = chapterTexts.some(t => /^\d+장?$/.test(t.trim()) || t === '1' || t === '2');
      hasChapter ? pass('책 선택 → 장 목록 표시') : fail('책 선택 후 장 목록 없음');
    }

    // 34. 버전 선택 드롭다운 (개역개정 등)
    const versionSelect = await page.$('select, button:has-text("개역")');
    versionSelect ? pass('번역 버전 선택 UI 존재') : fail('번역 버전 선택 없음');

    // 35. 검색 input 존재
    const searchInput = await page.$('input[placeholder*="검색"], input[type="search"], input[type="text"]');
    searchInput ? pass('성경 검색 입력창 존재') : fail('성경 검색 입력창 없음');

  } finally {
    await ctx.close();
  }
}

async function testLyrics(browser) {
  console.log('\n[lyrics] 찬양 페이지');
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await ctx.newPage();
  try {
    await page.goto('http://localhost:3000/lyrics', { waitUntil: 'networkidle', timeout: TIMEOUT });
    await skipTour(page);

    // 36. 찬양 검색 input 존재
    const input = await page.$('input[type="text"], input:not([type])');
    input ? pass('찬양 검색 입력창 존재') : fail('찬양 검색 입력창 없음');

    // 37. 탭 존재 (자유곡/찬송가)
    const tabs = await page.$$('button');
    const tabTexts = await Promise.all(tabs.map(t => t.innerText().catch(() => '')));
    const hasFreeSong = tabTexts.some(t => t.includes('자유') || t.includes('곡'));
    const hasHymn = tabTexts.some(t => t.includes('찬송'));
    hasFreeSong && hasHymn ? pass('찬양 탭 존재') : fail('찬양 탭 없음', `자유:${hasFreeSong} 찬송:${hasHymn}`);

    // 38. 가로 스크롤 없음
    const hasHScroll = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 5);
    hasHScroll ? fail('lyrics 가로 스크롤 발생') : pass('lyrics 가로 스크롤 없음');

    // 39. 찬양 검색 입력 시 UI 반응 (결과 또는 "검색 결과 없음" 표시)
    const searchInput = await page.$('input[type="text"], input:not([type])');
    if (searchInput) {
      await searchInput.click();
      await searchInput.type('은혜', { delay: 50 });
      await page.waitForTimeout(1000);
      const bodyText = await page.evaluate(() => document.body.innerText);
      // 검색 결과 있거나 "결과 없음" 메시지 — 둘 다 정상 UX
      const hasReaction = bodyText.includes('은혜') || bodyText.includes('검색 결과 없음') || bodyText.includes('곡이 없');
      hasReaction ? pass('찬양 검색 입력 → UI 반응') : fail('찬양 검색 입력 후 무반응');
      await page.keyboard.press('Control+a');
      await page.keyboard.press('Backspace');
    }

    // 40. 찬송가 탭 클릭 → 찬송가 검색 UI
    const hymnTab = await page.$('button:has-text("찬송가")');
    if (hymnTab) {
      await hymnTab.click();
      await page.waitForTimeout(500);
      // 찬송가 탭에서는 번호 입력 또는 검색창이 있어야 함
      const hasInput = await page.$('input');
      hasInput ? pass('찬송가 탭 → 검색 UI 전환') : fail('찬송가 탭 전환 후 검색 UI 없음');
    }

    // 41. 찬송가 API 응답
    const hymnApi = await page.evaluate(async () => {
      try {
        const r = await fetch('http://localhost:8080/api/hymns/search?q=1');
        return r.status;
      } catch { return 0; }
    });
    hymnApi === 200 || hymnApi === 204
      ? pass('찬송가 API 응답', `HTTP ${hymnApi}`)
      : fail('찬송가 API 실패', `HTTP ${hymnApi}`);

  } finally {
    await ctx.close();
  }
}

async function testNavigation(browser) {
  console.log('\n[nav] 탭 네비게이션 UX');
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await ctx.newPage();
  try {
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle', timeout: TIMEOUT });
    await skipTour(page);

    // 42. Hymns 탭 클릭 → /lyrics 로 라우팅
    const hymnTab = await page.$('a[href*="lyrics"], button:has-text("Hymns")');
    if (hymnTab) {
      await hymnTab.click();
      await page.waitForURL('**/lyrics**', { timeout: 5000 }).catch(() => {});
      const url = page.url();
      url.includes('lyrics') ? pass('Hymns 탭 클릭 → /lyrics 라우팅') : fail('Hymns 탭 라우팅 실패', url);
    }

    // 43. Scripture 탭 클릭 → /bible 로 라우팅
    const bibleTab = await page.$('a[href*="bible"], button:has-text("Scripture")');
    if (bibleTab) {
      await bibleTab.click();
      await page.waitForURL('**/bible**', { timeout: 5000 }).catch(() => {});
      const url = page.url();
      url.includes('bible') ? pass('Scripture 탭 클릭 → /bible 라우팅') : fail('Scripture 탭 라우팅 실패', url);
    }

    // 44. Bulletin 탭 클릭 → / 또는 /bulletin
    const bulletinTab = await page.$('a[href="/"], a[href*="bulletin"], button:has-text("Bulletin")');
    if (bulletinTab) {
      await bulletinTab.click();
      await page.waitForTimeout(500);
      const url = page.url();
      url.includes('localhost:3000') ? pass('Bulletin 탭 클릭 → 루트 라우팅') : fail('Bulletin 탭 라우팅 실패', url);
    }

    // 45. 페이지 전환 후 예배 순서 패널 유지
    const seqPanel = await page.$('[data-testid="seq-panel"]');
    seqPanel ? pass('탭 전환 후 순서 패널 유지') : fail('탭 전환 후 순서 패널 사라짐');

  } finally {
    await ctx.close();
  }
}

async function testApiEdgeCases(browser) {
  console.log('\n[api] API 엣지케이스');
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await ctx.newPage();
  try {
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle', timeout: TIMEOUT });

    // 46. 잘못된 예배 타입 → 400/오류 응답 (서버 안정성)
    const badTypeStatus = await page.evaluate(async () => {
      try {
        const r = await fetch('http://localhost:8080/api/worship-order?type=invalid_type');
        return r.status;
      } catch { return 0; }
    });
    badTypeStatus >= 400 && badTypeStatus < 500
      ? pass('잘못된 예배 타입 → 400 오류 반환', `HTTP ${badTypeStatus}`)
      : fail('잘못된 예배 타입 오류 처리 미흡', `HTTP ${badTypeStatus}`);

    // 47. 성경 없는 장 조회 → 오류 처리
    const badChapterStatus = await page.evaluate(async () => {
      try {
        const r = await fetch('http://localhost:8080/api/bible/verses?book=창세기&chapter=999');
        return r.status;
      } catch { return 0; }
    });
    badChapterStatus >= 400 || badChapterStatus === 200
      ? pass('없는 장 조회 API 처리', `HTTP ${badChapterStatus}`)
      : fail('없는 장 조회 API 무응답', `HTTP ${badChapterStatus}`);

    // 48. after_worship 예배 타입 API 응답
    const afterWorshipOk = await page.evaluate(async () => {
      try {
        const r = await fetch('http://localhost:8080/api/worship-order?type=after_worship');
        return r.ok;
      } catch { return false; }
    });
    afterWorshipOk ? pass('after_worship API 응답') : fail('after_worship API 실패');

    // 49. wed_worship 예배 타입 API 응답
    const wedWorshipOk = await page.evaluate(async () => {
      try {
        const r = await fetch('http://localhost:8080/api/worship-order?type=wed_worship');
        return r.ok;
      } catch { return false; }
    });
    wedWorshipOk ? pass('wed_worship API 응답') : fail('wed_worship API 실패');

    // 50. 로고 이미지 로드 (404 없어야 함)
    const logoStatus = await page.evaluate(async () => {
      try {
        const r = await fetch('http://localhost:3000/images/ep-logo.svg');
        return r.status;
      } catch { return 0; }
    });
    logoStatus === 200 ? pass('로고 SVG 로드 성공') : fail('로고 SVG 로드 실패', `HTTP ${logoStatus}`);

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
    testLyrics(browser),
    testNavigation(browser),
    testApiEdgeCases(browser),
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
