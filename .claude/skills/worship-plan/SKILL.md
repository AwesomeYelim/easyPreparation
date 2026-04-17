---
name: worship-plan
description: 이번 주 예배 준비 상태 분석 및 누락 항목 리포트. "예배 준비 확인", "에셋 확인", "찬송 캐시" 등 요청 시 자동 호출.
context: fork
agent: Explore
---

## 현재 예배 순서
!`cat config/main_worship.json 2>/dev/null | head -150 || echo "파일 없음 — config/*.json 탐색 필요"`

## 찬송 캐시 현황
!`ls data/hymn/ 2>/dev/null | head -50 || echo "캐시 없음"`

## 교독 캐시 현황
!`ls data/responsive_reading/ 2>/dev/null | head -50 || echo "캐시 없음"`

## PNG 캐시 현황
!`ls data/cache/hymn_pages/ 2>/dev/null | wc -l | xargs -I{} echo "{}개 파일"`

---

위 데이터를 바탕으로 이번 주 예배 준비 상태를 분석하세요:

1. **순서 파악**: 예배 순서 JSON에서 찬송/교독 번호 추출
2. **에셋 대조**: 캐시 현황과 비교 → 누락 항목 식별
3. **정보 완성도**: `info` 필드 누락 확인 (기도자, 성경 본문, 특순 제목 등)

출력 형식:
```
=== 예배 준비 리포트 ===
✅ 준비 완료: N개
⚠️  누락 에셋: [찬송 NNN, 교독 NN, ...]  ← 서버 재시작 시 자동 다운로드
❌ 정보 필요: [항목명: 이유]
권장 조치: [구체적 다음 단계]
```

추가 요청: $ARGUMENTS
