package godog

import (
	"fmt"
	"bytes"
	"strings"
	"testing"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/DATA-DOG/godog/colors"
)

func TestTeamcityFormatterOutput(t *testing.T) {
	feat, err := gherkin.ParseFeature(strings.NewReader(sampleGherkinFeature))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt: teamcityFunc("teamcity", w),
		features: []*feature{&feature{
			Path:    "any.feature",
			Feature: feat,
			Content: []byte(sampleGherkinFeature),
		}},
		initializer: func(s *Suite) {
			s.Step(`^passing$`, func() error { return nil })
			s.Step(`^failing$`, func() error { return fmt.Errorf("errored") })
			s.Step(`^pending$`, func() error { return ErrPending })
		},
	}

	expected := `

}`

	expected = trimAllLines(expected)

	r.run()

	actual := trimAllLines(buf.String())

	shouldMatchOutput(expected, actual, t)
}