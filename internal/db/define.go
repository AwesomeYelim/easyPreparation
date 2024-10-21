package db

// Song 구조체 정의 (Bolthold 데이터 저장용)
type Song struct {
	ID      int `boltholdKey:"ID"` // Primary Key 역할
	Title   string
	Content []string
}
