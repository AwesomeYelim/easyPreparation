import { test, expect } from '@playwright/test';

/*
 * Smoke E2E Tests
 * 전제: `make dev` 로 Go(:8080) + Next.js(:3000) 가 실행 중이어야 함
 * 서버가 내려가 있으면 전체 스킵
 */

const NEXT_URL = 'http://localhost:3000';
const GO_URL = 'http://localhost:8080';

/* ---------- 서버 연결 확인 ---------- */

async function isServerUp(url: string): Promise<boolean> {
  try {
    const res = await fetch(url, { signal: AbortSignal.timeout(3_000) });
    return res.ok || res.status < 500;
  } catch {
    return false;
  }
}

let nextUp = false;
let goUp = false;

test.beforeAll(async () => {
  [nextUp, goUp] = await Promise.all([
    isServerUp(NEXT_URL),
    isServerUp(GO_URL),
  ]);
});

/* ====================================
 * 1. Next.js 메인 페이지 로드
 * ==================================== */

test.describe('Next.js 메인 페이지', () => {
  test.beforeEach(() => {
    test.skip(!nextUp, 'Next.js 서버(localhost:3000)가 실행 중이 아닙니다');
  });

  test('페이지가 로드되고 주요 UI 요소가 보인다', async ({ page }) => {
    await page.goto('/');
    // 페이지 타이틀 또는 body 가 존재하는지 확인
    await expect(page.locator('body')).toBeVisible();
    // Next.js 앱이 마운트되면 #__next 가 존재
    await expect(page.locator('#__next')).toBeAttached();
  });
});

/* ====================================
 * 2. Display 페이지 렌더링 (Go 서버)
 * ==================================== */

test.describe('Display 페이지', () => {
  test.beforeEach(() => {
    test.skip(!goUp, 'Go 서버(localhost:8080)가 실행 중이 아닙니다');
  });

  test('/display 가 렌더링된다', async ({ page }) => {
    await page.goto(`${GO_URL}/display`);
    // Display HTML 이 정상 로드되면 body 가 존재
    await expect(page.locator('body')).toBeVisible();
    // 페이지에 에러 텍스트가 없는지 확인
    const bodyText = await page.locator('body').textContent();
    expect(bodyText).not.toContain('Internal Server Error');
  });

  test('/display/overlay 가 렌더링된다', async ({ page }) => {
    await page.goto(`${GO_URL}/display/overlay`);
    await expect(page.locator('body')).toBeVisible();
  });

  test('/display/stage 가 렌더링된다', async ({ page }) => {
    await page.goto(`${GO_URL}/display/stage`);
    await expect(page.locator('body')).toBeVisible();
  });
});

/* ====================================
 * 3. 헬스체크 API
 * ==================================== */

test.describe('헬스체크 API', () => {
  test.beforeEach(() => {
    test.skip(!goUp, 'Go 서버(localhost:8080)가 실행 중이 아닙니다');
  });

  test('/api/health 가 unhealthy 가 아니다', async ({ request }) => {
    const res = await request.get(`${GO_URL}/api/health`);
    expect(res.ok()).toBeTruthy();

    const body = await res.text();
    // JSON 응답이면 status 필드 확인
    try {
      const json = JSON.parse(body);
      expect(json.status).not.toBe('unhealthy');
    } catch {
      // JSON 이 아닌 텍스트 응답도 허용 (단순 "ok" 등)
      expect(body.toLowerCase()).not.toContain('unhealthy');
    }
  });

  test('/api/display-config 가 응답한다', async ({ request }) => {
    const res = await request.get(`${GO_URL}/api/display-config`);
    expect(res.ok()).toBeTruthy();
  });
});
