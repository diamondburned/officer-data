package acmcsuf

import (
	"fmt"
	"strconv"
)

// Constants for paths that point to the JSON files from the acmcsuf.com
// repository.
const (
	OfficersJSONPath = "./src/lib/public/board/data/officers.json"
	TiersJSONPath    = "./src/lib/public/board/data/tiers.json"
)

// Officers returns the officers for the given term.
type Officers []Officer

// Find returns the officer with the given name.
func (o Officers) Find(f func(*Officer) bool) *Officer {
	for i := range o {
		if f(&o[i]) {
			return &o[i]
		}
	}
	return nil
}

// Officer represents an ACM officer.
type Officer struct {
	FullName string               `json:"fullName"`
	Picture  string               `json:"picture"`
	Socials  Socials              `json:"socials"`
	Terms    map[Term]OfficerTerm `json:"terms"`
}

// Socials is an object of social media platforms to their URLs.
type Socials struct {
	Website   string `json:"website"`
	GitHub    string `json:"github"`
	Discord   string `json:"discord"`
	LinkedIn  string `json:"linkedin"`
	Instagram string `json:"instagram"`
}

// OfficerTerm represents a term of an officer.
type OfficerTerm struct {
	Title string // value in OfficerTiers
	Tier  int    // index in OfficerTiers
}

// Tiers is a list of known officer tiers, which is all the officer positions
// used for the order (sort of like a pyramid of hierarchy).
//
// The usefulness of this type and the file (tiers.json) is debatable.
type Tiers []string

// Term represents a term of an officer.
type Term string

// NewTerm returns a new term from the given semester and year.
func NewTerm(semester Semester, year int) Term {
	return Term(fmt.Sprintf("%s%d", semester, year))
}

// Semester returns the semester of the term.
func (t Term) Semester() Semester {
	if len(t) < 1 {
		return ""
	}
	return Semester(t[:1])
}

// Year returns the year of the term.
func (t Term) Year() int {
	if len(t) < 2 {
		return 0
	}

	year, err := strconv.Atoi(string(t[1:]))
	if err != nil {
		return 0
	}

	return year
}

// Validate returns an error if the term is invalid.
func (t Term) Validate() error {
	if t.Semester() == "" {
		return fmt.Errorf("invalid semester: %q", t)
	}

	if t.Year() == 0 {
		return fmt.Errorf("invalid year: %q", t)
	}

	return nil
}

// Semester represents a semester.
type Semester string

const (
	Fall   Semester = "F"
	Spring Semester = "S"
)

// String returns the human representation of the term.
func (s Semester) String() string {
	switch s {
	case Fall:
		return "Fall"
	case Spring:
		return "Spring"
	default:
		return fmt.Sprintf("Semester(%q)", string(s))
	}
}
