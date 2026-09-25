package result

import (
	"fmt"
	"time"
)

type Subject struct {
	Name  string `json:"subject"`
	Marks int    `json:"marks"`
	Grade string `json:"grade"`
}

type Result struct {
	Exam            string    `json:"exam"`
	ExamYear        int16     `json:"exam_year"`
	Board           string    `json:"board"`
	Roll            int32     `json:"roll"`
	Registration    int64     `json:"registration"`
	Name            string    `json:"name"`
	Division        string    `json:"division"`
	District        string    `json:"district"`
	InstitutionCode string    `json:"institution_code"`
	GPA             float64   `json:"gpa"`
	Passed          bool      `json:"passed"`
	Subjects        []Subject `json:"subjects"`
}

func (r Result) Key() Key {
	return Key{
		Exam: r.Exam, ExamYear: r.ExamYear, Board: r.Board,
		Roll: r.Roll, Registration: r.Registration,
	}
}

type Key struct {
	Exam         string
	ExamYear     int16
	Board        string
	Roll         int32
	Registration int64
}

func (k Key) rollPrefix() string {
	rollStr := fmt.Sprintf("%d", k.Roll)
	if len(rollStr) > 3 {
		return rollStr[:3]
	}
	return rollStr
}

func (k Key) PDFPath() string {
	prefix := k.rollPrefix()
	if len(prefix) > 3 {
		prefix = prefix[:3]
	}
	return fmt.Sprintf("%s/%d/%s/%s-%d-%d-%d.pdf",
		k.Board, k.ExamYear, prefix, k.Board, k.ExamYear, k.Roll, k.Registration)
}

func (k Key) ResultPath() string {
	return fmt.Sprintf("%s/%d/%s/%s-%d-%d-%d.html",
		k.Board, k.ExamYear, k.rollPrefix(), k.Board, k.ExamYear, k.Roll, k.Registration)
}

func (k Key) PDFURL() string    { return "/marksheets/" + k.PDFPath() }
func (k Key) ResultURL() string { return "/results/" + k.ResultPath() }

type BoardStat struct {
	Board    string  `json:"board"`
	Total    int64   `json:"total"`
	Passed   int64   `json:"passed"`
	Failed   int64   `json:"failed"`
	PassRate float64 `json:"pass_rate"`
	AvgGPA   float64 `json:"avg_gpa"`
}

// Stats is the aggregate result summary for one exam/year.
type Stats struct {
	Exam        string      `json:"exam"`
	ExamYear    int16       `json:"exam_year"`
	Total       int64       `json:"total"`
	Passed      int64       `json:"passed"`
	Failed      int64       `json:"failed"`
	PassRate    float64     `json:"pass_rate"`
	AvgGPA      float64     `json:"avg_gpa"`
	ByBoard     []BoardStat `json:"by_board"`
	GeneratedAt time.Time   `json:"generated_at"`
}