package handlers

import (
	"testing"
	"time"
)

func TestDueScheduleStep(t *testing.T) {
	loc := time.Local
	// 2026-11-08 은 일요일
	sunday := func(h, m, s int) time.Time { return time.Date(2026, 11, 8, h, m, s, 0, loc) }
	entry := ScheduleEntry{WorshipType: "main_worship", Label: "주일예배", Weekday: 0, Hour: 11, Minute: 0, Enabled: true}

	cases := []struct {
		name      string
		entry     ScheduleEntry
		now       time.Time
		countdown int
		want      scheduleStep
	}{
		{"카운트다운 전", entry, sunday(10, 54, 59), 5, scheduleStep{}},
		{"카운트다운 시작 = 송출 준비", entry, sunday(10, 55, 0), 5, scheduleStep{countdown: 300, prep: true}},
		{"카운트다운 중", entry, sunday(10, 59, 30), 5, scheduleStep{countdown: 30, prep: true}},
		{"예배 시각 = 순서 시작", entry, sunday(11, 0, 0), 5, scheduleStep{prep: true, start: true}},
		{"시각을 놓쳤지만 유예 안", entry, sunday(11, 4, 59), 5, scheduleStep{prep: true, start: true}},
		{"유예 끝", entry, sunday(11, 5, 1), 5, scheduleStep{}},
		{"다른 요일", entry, sunday(11, 0, 0).AddDate(0, 0, 1), 5, scheduleStep{}},
		{"꺼진 항목", ScheduleEntry{Weekday: 0, Hour: 11, Enabled: false}, sunday(11, 0, 0), 5, scheduleStep{}},
		{"카운트다운 0분이면 예배 시각에 둘 다", entry, sunday(10, 59, 59), 0, scheduleStep{}},
		{"카운트다운 0분, 예배 시각", entry, sunday(11, 0, 0), 0, scheduleStep{prep: true, start: true}},
		{"카운트다운 마지막 1초 미만", entry, sunday(10, 59, 59).Add(500 * time.Millisecond), 5, scheduleStep{countdown: 1, prep: true}},
	}
	for _, c := range cases {
		if got := dueScheduleStep(c.entry, c.now, c.countdown); got != c.want {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestOperatorActiveWithin(t *testing.T) {
	orderMu.Lock()
	saved := lastOperatorActivity
	lastOperatorActivity = time.Time{}
	orderMu.Unlock()
	defer func() {
		orderMu.Lock()
		lastOperatorActivity = saved
		orderMu.Unlock()
	}()

	if operatorActiveWithin(time.Hour) {
		t.Error("활동 기록이 없으면 진행 중이 아니어야 한다 (재시작 후 복원 상태)")
	}
	orderMu.Lock()
	lastOperatorActivity = time.Now().Add(-10 * time.Minute)
	orderMu.Unlock()
	if !operatorActiveWithin(time.Hour) {
		t.Error("10분 전 활동은 진행 중으로 봐야 한다")
	}
	orderMu.Lock()
	lastOperatorActivity = time.Now().Add(-2 * time.Hour)
	orderMu.Unlock()
	if operatorActiveWithin(time.Hour) {
		t.Error("2시간 전 활동은 진행 중이 아니어야 한다")
	}
}
