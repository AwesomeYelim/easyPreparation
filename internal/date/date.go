package date

import (
	"fmt"
	"time"
)

// 파일명용 date 날짜
// 202411, 3 => 2024년 11월 3째주
func SetDateTitle() (string, string) {
	currentDate := time.Now()
	firstDayOfMonth := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())
	_, firstWeek := firstDayOfMonth.ISOWeek()
	_, currentWeek := currentDate.ISOWeek()
	weekInMonth := currentWeek - firstWeek + 1
	yearMonth := currentDate.Format("200601")
	weekFormatted := fmt.Sprintf("%d", weekInMonth)

	return yearMonth, weekFormatted
}

// 이번주 주일 날짜 계산
func SetThisSunDay() string {
	currentDate := time.Now()
	daysUntilSunday := (7 - int(currentDate.Weekday())) % 7
	thisSunday := currentDate.AddDate(0, 0, daysUntilSunday)

	dateText := thisSunday.Format("2006년 01월 02일")

	return dateText
}
