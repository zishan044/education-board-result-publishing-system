package result

import "fmt"

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

type Key struct {
	Exam         string
	ExamYear     int16
	Board        string
	Roll         int32
	Registration int64
}

func (k Key) PDFPath() string {
	rollStr := fmt.Sprintf("%d", k.Roll)
	prefix := rollStr
	if len(prefix) > 3 {
		prefix = prefix[:3]
	}
	return fmt.Sprintf("%s/%d/%s/%s-%d-%d-%d.pdf",
		k.Board, k.ExamYear, prefix, k.Board, k.ExamYear, k.Roll, k.Registration)
}