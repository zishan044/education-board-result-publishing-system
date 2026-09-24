package pdfgen

import (
	"fmt"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"github.com/zishan044/education-board-result-publishing-system/internal/result"
)

func Generate(r *result.Result) ([]byte, error) {
	cfg := config.NewBuilder().
		WithPageNumber().
		Build()

	m := maroto.New(cfg)

	m.AddRows(
		row.New(20).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("%s Examination %d", r.Exam, r.ExamYear),
					props.Text{Size: 16, Style: fontstyle.Bold, Align: align.Center}),
			),
		),
		row.New(8).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Education Board: %s", r.Board),
					props.Text{Size: 11, Align: align.Center}),
			),
		),
	)

	m.AddRows(row.New(6))

	infoRows := [][2]string{
		{"Name", r.Name},
		{"Roll", fmt.Sprintf("%d", r.Roll)},
		{"Registration", fmt.Sprintf("%d", r.Registration)},
		{"Institution Code", r.InstitutionCode},
		{"District", r.District},
		{"Division", r.Division},
	}
	for _, kv := range infoRows {
		m.AddRows(
			row.New(7).Add(
				col.New(4).Add(text.New(kv[0], props.Text{Size: 10, Style: fontstyle.Bold})),
				col.New(8).Add(text.New(kv[1], props.Text{Size: 10})),
			),
		)
	}

	m.AddRows(row.New(6))

	m.AddRows(
		row.New(8).Add(
			col.New(6).Add(text.New("Subject", props.Text{Size: 10, Style: fontstyle.Bold})),
			col.New(3).Add(text.New("Marks", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Center})),
			col.New(3).Add(text.New("Grade", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Center})),
		),
	)
	for _, s := range r.Subjects {
		m.AddRows(
			row.New(7).Add(
				col.New(6).Add(text.New(s.Name, props.Text{Size: 10})),
				col.New(3).Add(text.New(fmt.Sprintf("%d", s.Marks), props.Text{Size: 10, Align: align.Center})),
				col.New(3).Add(text.New(s.Grade, props.Text{Size: 10, Align: align.Center})),
			),
		)
	}

	m.AddRows(row.New(6))

	result := "PASSED"
	if !r.Passed {
		result = "FAILED"
	}
	m.AddRows(
		row.New(10).Add(
			col.New(6).Add(text.New(fmt.Sprintf("GPA: %.2f", r.GPA),
				props.Text{Size: 12, Style: fontstyle.Bold})),
			col.New(6).Add(text.New(result,
				props.Text{Size: 12, Style: fontstyle.Bold, Align: align.Right})),
		),
	)

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("render pdf: %w", err)
	}
	return doc.GetBytes(), nil
}