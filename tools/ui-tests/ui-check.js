#!/usr/bin/env node
// UI 통합 체크 — 스크린샷 + 기능 테스트 병렬 실행, 실시간 출력
// 사용: make ui-check  (프로젝트 루트)
//       cd ui && npm run ui:check

const { spawn } = require('child_process');
const path = require('path');

const R = '\x1b[0m';
const BOLD = '\x1b[1m';
const RED = '\x1b[31m';
const GREEN = '\x1b[32m';
const CYAN = '\x1b[36m';
const YELLOW = '\x1b[33m';
const DIM = '\x1b[2m';

function runScript(scriptName, tag, color) {
  return new Promise((resolve) => {
    const proc = spawn('node', [path.join(__dirname, scriptName)], {
      cwd: path.join(__dirname, '../ui'),
      stdio: ['ignore', 'pipe', 'pipe'],
      env: { ...process.env, NODE_PATH: path.join(__dirname, '../ui/node_modules') },
    });

    let ok = true;
    const failures = [];

    const onLine = (line, isErr) => {
      if (!line.trim()) return;
      const mark = isErr || line.includes('✗') ? `${RED}✗${R}` : '';
      process.stdout.write(`${color}${DIM}[${tag}]${R} ${line}\n`);
      if (isErr || line.includes('✗')) {
        ok = false;
        failures.push(line.trim());
      }
    };

    let stdoutBuf = '';
    proc.stdout.on('data', d => {
      stdoutBuf += d.toString();
      const parts = stdoutBuf.split('\n');
      stdoutBuf = parts.pop();
      parts.forEach(l => onLine(l, false));
    });

    let stderrBuf = '';
    proc.stderr.on('data', d => {
      stderrBuf += d.toString();
      const parts = stderrBuf.split('\n');
      stderrBuf = parts.pop();
      parts.forEach(l => onLine(l, true));
    });

    proc.on('close', code => {
      if (stdoutBuf.trim()) onLine(stdoutBuf, false);
      if (stderrBuf.trim()) onLine(stderrBuf, true);
      resolve({ tag, ok: code === 0 && ok, failures });
    });
  });
}

async function main() {
  console.log(`\n${BOLD}🔍 UI 체크 시작 (병렬)${R}\n`);

  const [shots, tests] = await Promise.all([
    runScript('ui-screenshot.js', '📸 스크린샷', CYAN),
    runScript('ui-test.js',       '🧪 기능',     YELLOW),
  ]);

  console.log(`\n${DIM}${'─'.repeat(48)}${R}`);

  const lines = [shots, tests].map(r => {
    const icon = r.ok ? `${GREEN}✅ 통과${R}` : `${RED}❌ 실패${R}`;
    let out = `${r.tag}: ${icon}`;
    if (!r.ok && r.failures.length) {
      r.failures.slice(0, 5).forEach(f => { out += `\n  ${RED}→${R} ${f}`; });
    }
    return out;
  });

  lines.forEach(l => console.log(l));

  if (shots.ok && tests.ok) {
    console.log(`\n${GREEN}${BOLD}✅ 모든 체크 통과${R}\n`);
  } else {
    console.log(`\n${RED}${BOLD}❌ 버그 발견 — Claude에게 "UI 디버그 돌려줘" 로 자동 수정${R}\n`);
    process.exit(1);
  }
}

main().catch(e => { console.error(e.message); process.exit(1); });
