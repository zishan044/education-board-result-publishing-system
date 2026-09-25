package render

import (
	"embed"
	"html/template"
	"io"

	"github.com/zishan044/education-board-result-publishing-system/internal/result"
)

var templateFS embed.FS

var funcs = template.FuncMap{
	"gpaClass": func(gpa float64) string {
		switch {
		case gpa >= 5.0:
			return "gpa-perfect"
		case gpa >= 3.5:
			return "gpa-good"
		case gpa > 0:
			return "gpa-pass"
		default:
			return "gpa-fail"
		}
	},
	"pdfURL":    func(r *result.Result) string { return r.Key().PDFURL() },
	"resultURL": func(r *result.Result) string { return r.Key().ResultURL() },
}

var (
	resultTmpl = template.Must(template.New("layout.html").Funcs(funcs).
			ParseFS(templateFS, "templates/layout.html", "templates/result.html"))
	statsTmpl = template.Must(template.New("layout.html").Funcs(funcs).
			ParseFS(templateFS, "templates/layout.html", "templates/stats.html"))
	indexTmpl = template.Must(template.New("layout.html").Funcs(funcs).
			ParseFS(templateFS, "templates/layout.html", "templates/index.html"))
)

func Result(w io.Writer, r *result.Result) error {
	return resultTmpl.ExecuteTemplate(w, "layout.html", r)
}

func Stats(w io.Writer, s *result.Stats) error {
	return statsTmpl.ExecuteTemplate(w, "layout.html", s)
}

type IndexData struct {
	Boards []string
	Years  []int16
}

func Index(w io.Writer, d IndexData) error {
	return indexTmpl.ExecuteTemplate(w, "layout.html", d)
}