/**
 * Go 서버 에러 메시지 → 사용자 친화적 한국어 변환
 */
export function translateError(msg: string): string {
  const map: [RegExp | string, string][] = [
    ["OBS connection", "OBS 연결 실패 — OBS가 실행 중이고 WebSocket 플러그인이 활성화되어 있는지 확인하세요"],
    ["obs", "OBS 연결 실패 — OBS가 실행 중이고 WebSocket 플러그인이 활성화되어 있는지 확인하세요"],
    ["bible DB", "성경 DB 파일이 없습니다 — 앱을 재시작해 주세요"],
    ["bible.db", "성경 DB 파일이 없습니다 — 앱을 재시작해 주세요"],
    ["gs 변환 실패", "PDF 변환 실패 — Ghostscript 설치가 필요합니다"],
    ["ghostscript", "PDF 변환 실패 — Ghostscript 설치가 필요합니다"],
    [/gs(win64c|win32c)? not found/i, "PDF 변환 실패 — Ghostscript 설치가 필요합니다"],
    ["파일 저장 실패", "파일 저장에 실패했습니다 — 디스크 용량을 확인해 주세요"],
    ["failed to write", "파일 저장에 실패했습니다 — 디스크 용량을 확인해 주세요"],
    ["업로드 실패", "업로드에 실패했습니다 — 파일 크기와 형식을 확인해 주세요"],
    ["upload failed", "업로드에 실패했습니다 — 파일 크기와 형식을 확인해 주세요"],
    ["네트워크", "네트워크 오류가 발생했습니다 — 인터넷 연결을 확인해 주세요"],
    [/network error/i, "네트워크 오류가 발생했습니다 — 인터넷 연결을 확인해 주세요"],
    [/connection refused/i, "서버에 연결할 수 없습니다 — 서버가 실행 중인지 확인해 주세요"],
    [/timeout/i, "요청 시간이 초과되었습니다 — 잠시 후 다시 시도해 주세요"],
    [/permission denied/i, "파일 접근 권한이 없습니다 — 관리자 권한으로 실행해 주세요"],
    ["license", "라이선스 확인 실패 — 라이선스 키를 확인해 주세요"],
    [/pdf.*error/i, "PDF 처리 중 오류가 발생했습니다"],
    [/database|db error/i, "데이터베이스 오류가 발생했습니다 — 앱을 재시작해 주세요"],
  ];

  const lower = msg.toLowerCase();
  for (const [pattern, korean] of map) {
    if (typeof pattern === "string") {
      if (lower.includes(pattern.toLowerCase())) return korean;
    } else {
      if (pattern.test(msg)) return korean;
    }
  }
  return msg;
}

/**
 * progress 메시지 코드
 * -1: 에러
 *  0: 진행 중
 *  1: 완료
 */
export const PROGRESS_ERROR = -1;
export const PROGRESS_IN_PROGRESS = 0;
export const PROGRESS_DONE = 1;

/**
 * 완료 메시지에서 toast.success로 표시할 중요 완료 패턴
 * (모든 완료 메시지를 toast로 띄우면 시끄러우므로 중요한 것만 선별)
 */
export function isImportantSuccess(message: string): boolean {
  const patterns = [
    "완료",
    "생성됨",
    "저장됨",
    "업로드",
    "활성화",
    "연결",
    "generated",
    "uploaded",
    "saved",
    "connected",
  ];
  const lower = message.toLowerCase();
  return patterns.some((p) => lower.includes(p.toLowerCase()));
}
