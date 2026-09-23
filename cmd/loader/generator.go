package main

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"slices"
	"sort"

	"github.com/zishan044/education-board-result-publishing-system/internal/result"
)

var resultColumns = []string{
	"exam", "exam_year", "board", "roll", "registration", "name",
	"division", "district", "institution_code", "gpa", "passed", "subjects",
}

type divisionInfo struct {
	Name      string
	Board     string
	Weight    float64
	Districts []string
}

var divisions = []divisionInfo{
	{"Dhaka", "Dhaka", 30, []string{"Dhaka", "Gazipur", "Narayanganj", "Tangail", "Kishoreganj", "Narsingdi", "Faridpur", "Manikganj", "Munshiganj", "Gopalganj", "Madaripur", "Rajbari", "Shariatpur"}},
	{"Chattogram", "Chattogram", 18, []string{"Chattogram", "Cumilla", "Noakhali", "Cox's Bazar", "Brahmanbaria", "Chandpur", "Feni", "Lakshmipur", "Rangamati", "Khagrachhari", "Bandarban"}},
	{"Rajshahi", "Rajshahi", 10, []string{"Rajshahi", "Bogura", "Pabna", "Sirajganj", "Naogaon", "Natore", "Chapainawabganj", "Joypurhat"}},
	{"Khulna", "Jashore", 9, []string{"Khulna", "Jashore", "Kushtia", "Satkhira", "Bagerhat", "Jhenaidah", "Magura", "Chuadanga", "Narail", "Meherpur"}},
	{"Barishal", "Barishal", 5, []string{"Barishal", "Patuakhali", "Bhola", "Pirojpur", "Jhalokati", "Barguna"}},
	{"Sylhet", "Sylhet", 6, []string{"Sylhet", "Moulvibazar", "Habiganj", "Sunamganj"}},
	{"Rangpur", "Dinajpur", 11, []string{"Rangpur", "Dinajpur", "Gaibandha", "Kurigram", "Nilphamari", "Thakurgaon", "Lalmonirhat", "Panchagarh"}},
	{"Mymensingh", "Mymensingh", 8, []string{"Mymensingh", "Netrokona", "Jamalpur", "Sherpur"}},
}

type groupInfo struct {
	Name     string
	Weight   float64
	Subjects []string
}

var commonSubjects = []string{"Bangla", "English", "Mathematics", "Religion and Moral Education", "ICT"}

var groups = []groupInfo{
	{"Science", 40, []string{"Physics", "Chemistry", "Biology", "Higher Mathematics"}},
	{"Business Studies", 25, []string{"Accounting", "Business Entrepreneurship", "Finance and Banking"}},
	{"Humanities", 35, []string{"Geography and Environment", "Civics", "Economics"}},
}

var firstNames = []string{"Nirjhor", "Rahim", "Karim", "Fatima", "Ayesha", "Tanvir", "Sadia", "Imran", "Nusrat", "Mahmud",
	"Farhan", "Tasnim", "Rafiq", "Sumaiya", "Arif", "Jannat", "Shakib", "Mitu", "Habib", "Rumana"}

var lastNames = []string{"Ahmed", "Hossain", "Rahman", "Islam", "Khan", "Chowdhury", "Akter", "Begum",
	"Uddin", "Sarker", "Mia", "Ali", "Talukder", "Das", "Roy"}


type weighted struct {
	cum   []float64
	total float64
}

func newWeighted(ws []float64) weighted {
	cum := make([]float64, len(ws))
	var sum float64
	for i, w := range ws {
		sum += w
		cum[i] = sum
	}
	return weighted{cum: cum, total: sum}
}

func (w weighted) pick(r *rand.Rand) int {
	return sort.SearchFloat64s(w.cum, r.Float64()*w.total)
}


type Config struct {
	Exam        string
	Year        int16
	Count       int64
	Seed        uint64
	SampleEvery int64
}

type institution struct {
	code      string
	divIdx    int
	district  string
	remaining int
}

type Generator struct {
	cfg Config
	rng *rand.Rand

	divPick      weighted
	districtPick []weighted
	groupPick    weighted
	subjectLists [][]string

	inst      institution
	instCount int
	nextRoll  []int32
	produced  int64
	cur       result.Result
	sample    []result.Key
}

func NewGenerator(cfg Config) *Generator {
	divWeights := make([]float64, len(divisions))
	districtPick := make([]weighted, len(divisions))
	for i, d := range divisions {
		divWeights[i] = d.Weight
		dw := make([]float64, len(d.Districts))
		for j := range dw {
			dw[j] = 1 / float64(j+1)
		}
		districtPick[i] = newWeighted(dw)
	}

	groupWeights := make([]float64, len(groups))
	subjectLists := make([][]string, len(groups))
	for i, gr := range groups {
		groupWeights[i] = gr.Weight
		subjectLists[i] = append(slices.Clone(commonSubjects), gr.Subjects...)
	}

	return &Generator{
		cfg:          cfg,
		rng:          rand.New(rand.NewPCG(cfg.Seed, cfg.Seed^0x9e3779b97f4a7c15)),
		divPick:      newWeighted(divWeights),
		districtPick: districtPick,
		groupPick:    newWeighted(groupWeights),
		subjectLists: subjectLists,
		nextRoll:     make([]int32, len(divisions)),
	}
}

func (g *Generator) Next() bool {
	if g.produced >= g.cfg.Count {
		return false
	}
	if g.inst.remaining == 0 {
		g.newInstitution()
	}
	g.inst.remaining--
	g.buildRow()
	g.produced++
	if g.produced%1_000_000 == 0 {
		log.Printf("generated %d rows", g.produced)
	}
	return true
}

func (g *Generator) Values() ([]any, error) {
	r := &g.cur
	return []any{
		r.Exam, r.ExamYear, r.Board, r.Roll, r.Registration, r.Name,
		r.Division, r.District, r.InstitutionCode, r.GPA, r.Passed, r.Subjects,
	}, nil
}

func (g *Generator) Err() error { return nil }

func (g *Generator) Produced() int64     { return g.produced }
func (g *Generator) Sample() []result.Key { return g.sample }

func (g *Generator) newInstitution() {
	divIdx := g.divPick.pick(g.rng)
	div := divisions[divIdx]

	size := int(math.Exp(4.5 + 0.7*g.rng.NormFloat64()))
	size = min(max(size, 5), 1500)

	g.instCount++
	g.inst = institution{
		code:      fmt.Sprintf("%06d", g.instCount),
		divIdx:    divIdx,
		district:  div.Districts[g.districtPick[divIdx].pick(g.rng)],
		remaining: size,
	}
}

func (g *Generator) buildRow() {
	div := divisions[g.inst.divIdx]
	gi := g.groupPick.pick(g.rng)
	names := g.subjectLists[gi]

	g.nextRoll[g.inst.divIdx]++
	roll := 1_000_000 + g.nextRoll[g.inst.divIdx]
	registration := 2_000_000_000 + g.produced

	ability := g.rng.NormFloat64() * 9
	subjects := make([]result.Subject, len(names))
	var points float64
	failed := false
	for i, n := range names {
		m := int(math.Round(64 + ability + g.rng.NormFloat64()*12))
		m = min(max(m, 0), 100)
		grade, gp := gradeFor(m)
		if grade == "F" {
			failed = true
		}
		points += gp
		subjects[i] = result.Subject{Name: n, Marks: m, Grade: grade}
	}

	gpa := 0.0
	if !failed {
		gpa = math.Round(points/float64(len(names))*100) / 100
	}

	g.cur = result.Result{
		Exam:            g.cfg.Exam,
		ExamYear:        g.cfg.Year,
		Board:           div.Board,
		Roll:            roll,
		Registration:    registration,
		Name:            firstNames[g.rng.IntN(len(firstNames))] + " " + lastNames[g.rng.IntN(len(lastNames))],
		Division:        div.Name,
		District:        g.inst.district,
		InstitutionCode: g.inst.code,
		GPA:             gpa,
		Passed:          !failed,
		Subjects:        subjects,
	}

	if g.cfg.SampleEvery > 0 && g.produced%g.cfg.SampleEvery == 0 {
		g.sample = append(g.sample, result.Key{
			Exam: g.cfg.Exam, ExamYear: g.cfg.Year, Board: div.Board,
			Roll: roll, Registration: registration,
		})
	}
}

func gradeFor(marks int) (string, float64) {
	switch {
	case marks >= 80:
		return "A+", 5.0
	case marks >= 70:
		return "A", 4.0
	case marks >= 60:
		return "A-", 3.5
	case marks >= 50:
		return "B", 3.0
	case marks >= 40:
		return "C", 2.0
	case marks >= 33:
		return "D", 1.0
	default:
		return "F", 0.0
	}
}