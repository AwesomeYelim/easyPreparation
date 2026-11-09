# easyPreparation 운영·개발 허브 (2026-11-09 기준)

easyPreparation(교회 예배 준비·송출 앱) 의 운영 현황·결정·다음 할 일. 여기엔 결론만.
쇼츠 자동화와 유료화·수익 메모는 life 저장소(`notes/side/autoshorts.md`, `notes/finance/side-income.md`)에 있다.

## 세션 여는 법 — 이 파일이 맥락 원본
- 로컬·클라우드(폰) 어느 세션이든 이 저장소만 열면 된다. `CLAUDE.md` 가 이 파일을 불러온다.
- 세션을 끝낼 때 "현재 상태"·"진행 중 브랜치"·"다음 할 일" 을 고치고 `scripts/push-context.sh` 를 실행한다. 이 파일만 지금 브랜치에 커밋·push 하면, 매시간 가장 최근 것이 master 로 모인다.
- life 저장소 `notes/side/easyprep.md` 는 이 파일의 복사본이다(매시간 자동). 거기서 고치지 않는다.
- 코드 수정은 브랜치로. 운영 중이므로 수정 내용을 먼저 보여 주고 승인받는다.
- 클라우드 세션이 못 하는 것: 교회 PC·배포 서버 접속, 로컬 `.env`, Windows 실행·OBS 실측. "다음 할 일" 에 `(PC)` 를 붙여 남긴다.
- 공개 저장소다. 호스트명·키·개인 정보는 적지 않는다.

## 현재 상태
- 최신 정식 릴리스: v1.5.9 (2026-11-02, master `6223852`). 릴리스 노트 맨 위에 v1.5.8 이후 요약.
- 배포 서버(사설망): `bulletin.service` 가 릴리스 바이너리 `bin/easyPreparation_server` 실행, 데이터 `EASYPREP_DATA_DIR=/home/ubuntu/easyPreparation`.
  - `~/bin/ep-update`: GitHub releases/latest(beta 제외) 받아 checksums 검증 → 교체 → `/api/version` 확인, 실패 시 `.prev` 로 복구. cron `0 3 * * 1-6`(일요일 제외). 로그 `~/logs/ep-update.log`.
  - 옛 `deploy.yml`(master push 때 dev 빌드 업로드) 은 `gh workflow disable` 로 끔. 되돌리기는 enable.
- church-pc: 개발 모드(`make dev`) 로 운영 중으로 보임(오프라인이라 미확인). 개발 모드는 업데이트 배너·릴리스가 적용되지 않는다.

## 개발 모드 → 운영판 전환 (미착수, 사용자 선택 대기)
개발/운영 차이(코드 기준):
1. 라이선스: 개발 빌드는 게이트 우회(`internal/middleware/dev.go`). 운영 Free 플랜은 기능 0개라 OBS 제어 등이 403.
2. 화면: 개발은 Next dev 서버(:3000), 운영은 바이너리 내장 정적 파일. 상대경로 `/api` 동작 차이 위험.
3. 데이터: 개발은 저장소 폴더, 운영은 `%APPDATA%\easyPreparation`. `EASYPREP_DATA_DIR` 로 기존 폴더 지정 가능.
권하는 순서: 교회용 라이선스 발급(`make build-devlicense-windows`) → 데이터 경로 지정 → 배포 서버를 beta 시험장으로(ep-update 에 beta 옵션) → 토요일 리허설 → 몇 주간 개발 모드를 비상구로 유지.

## 진행 중 브랜치
- 없음 (세션이 브랜치를 만들면 `브랜치 — 목적 — 남은 일` 한 줄로 적고, master 에 병합되면 지운다)

## 다음 할 일
- [ ] church-pc 가 켜지면 실제 실행 방식(개발/운영) 확인
- [ ] 라이선스 발급 + beta 시험장 중 무엇부터 할지 사용자 결정
- [ ] (PC) 배포 서버 클라우드 계정에 남은 예전 부트 볼륨 1개 확인 후 정리 여부 결정

## 기록
- (11-02) 배포 서버를 릴리스 기반으로 전환(v1.5.8 → v1.5.9 자동 반영 예정), v1.5.9 정식 릴리스.
- (11-09) 상시 서버는 확보 불가. 프로젝트 맥락은 파일로 이어 가기로 결정.
- (11-09) 맥락 원본을 이 파일로 옮김. PC 를 꺼도 폰 클라우드 세션에서 이어 간다. life 노트는 복사본.
