package result

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