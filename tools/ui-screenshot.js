#!/usr/bin/env node
// UI 스크린샷 자동 캡처 — Playwright headless 병렬
// 실행: cd ui && npm run screenshot
// 전제: make dev 로 서버 실행 중 (Go :8080 + Next.js :3000)

const { chromium } = require('@playwright/test');
const path = require('path');
const fs = require('fs');

const OUT_DIR = path.join(__dirname, '../.claude/screenshots');

const VIEWS = [
  { url: 'http://localhost:3000',               name: 'main',          w: 1280, h: 800  },
  { url: 'http://localhost:3000',               name: 'main-mobile',   w: 375,  h: 812  },
  { url: 'http://localhost:3000/bible',         name: 'bible',         w: 1280, h: 800  },
  { url: 'http://localhost:3000/bulletin',      name: 'bulletin',      w: 1280, h: 800  },
  { url: 'http://localhost:3000/lyrics',        name: 'lyrics',        w: 1280, h: 800  },
  { url: 'http://localhost:8080/display',       name: 'display-obs',   w: 1920, h: 1080 },
  { url: 'http://localhost:8080/display/stage', name: 'display-stage', w: 1920, h: 1080 },
];

async function dismissModal(page) {
  try {
    // 온보딩 모달 "건너뛰기" 버튼 클릭
    const skip = await page.$('button:has-text("건너뛰기")');
    if (skip) {
      await skip.click();
      await page.waitForTimeout(400);
    }
  } catch {}
}

async function capture(browser, view) {
  const ctx = await browser.newContext({
    viewport: { width: view.w, height: view.h },
    // localStorage로 온보딩 완료 상태 주입 (Next.js 앱용)
    storageState: {
      cookies: [],
      origins: [{
        origin: 'http://localhost:3000',
        localStorage: [{ name: 'ep_tour_done', value: '1' }],
      }],
    },
  });
  const page = await ctx.newPage();
  try {
    await page.goto(view.url, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(500);
    // localStorage로 안 닫히면 버튼으로 닫기
    await dismissModal(page);
    await page.waitForTimeout(300);
    const file = path.join(OUT_DIR, `${view.name}.png`);
    await page.screenshot({ path: file, fullPage: false });
    console.log(`✓ ${view.name}`);
    return { name: view.name, ok: true };
  } catch (e) {
    console.error(`✗ ${view.name} (${view.url}): ${e.message}`);
    return { name: view.name, ok: false };
  } finally {
    await ctx.close();
  }
}

async function main() {
  fs.mkdirSync(OUT_DIR, { recursive: true });
  console.log(`스크린샷 캡처 중... (${VIEWS.length}개 병렬)`);

  const browser = await chromium.launch({ headless: true });
  const results = await Promise.all(VIEWS.map(v => capture(browser, v)));
  await browser.close();

  const failed = results.filter(r => !r.ok);
  if (failed.length) {
    console.error(`\n실패 ${failed.length}/${VIEWS.length}: ${failed.map(f => f.name).join(', ')}`);
    console.error('서버가 실행 중인지 확인: make dev');
    process.exit(1);
  }
  console.log(`\n완료: ${OUT_DIR}`);
}

main().catch(e => { console.error(e.message); process.exit(1); });
