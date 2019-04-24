package godog

import (
	"bytes"
	"time"
	"fmt"
	"io"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/acarl005/stripansi"
)

func init() {
	Format("teamcity", "Prints teamcity compatable output to stdout.", teamcityFunc)
}

func teamcityFunc(suite string, out io.Writer) Formatter {
	return newTeamcityFormatter(suite, out)
}

type teamcityFormatter struct {
	out io.Writer
	pretty Formatter
	split *splitWriter

	currentFeature *gherkin.Feature

	currentScenario *gherkin.Scenario
	scenarioStarted int64
	scenarioFailed bool
}

func newTeamcityFormatter(suite string, out io.Writer) *teamcityFormatter {
	split := newSplitWriter(out)
	pretty := prettyFunc(suite, split)

	return &teamcityFormatter{out:out, pretty:pretty, split:split}
}

func (f *teamcityFormatter) Feature(feature *gherkin.Feature, path string, c []byte) {
	f.checkAndCloseCurrentFeature()

	line := fmt.Sprintf("##teamcity[testSuiteStarted name='Feature: %s']", f.escape(feature.Name))
	f.printLine(line)
	f.currentFeature = feature

	f.pretty.Feature(feature, path, c)
}

func (f *teamcityFormatter) Node(n interface{}) {
	switch v := n.(type) {
	case *gherkin.Scenario:
		f.checkAndCloseCurrentScenario()

		line := fmt.Sprintf("##teamcity[testStarted name='Scenario %s' captureStandardOutput='false']", f.escape(v.Name))
		f.printLine(line)

		f.currentScenario = v
		f.scenarioStarted = time.Now().UnixNano() / int64(time.Millisecond)
		f.scenarioFailed = false
	}

	f.pretty.Node(n)
}

func (f *teamcityFormatter) Defined(s *gherkin.Step, d *StepDef) {
	f.pretty.Defined(s, d)
}

func (f *teamcityFormatter) Failed(s *gherkin.Step, d *StepDef, err error) {
	f.scenarioFailed = true

	f.pretty.Failed(s, d, err)
}

func (f *teamcityFormatter) Passed(s *gherkin.Step, d *StepDef) {
	f.pretty.Passed(s, d)
}

func (f *teamcityFormatter) Skipped(s *gherkin.Step, d *StepDef) {
	f.pretty.Skipped(s, d)
}

func (f *teamcityFormatter) Undefined(s *gherkin.Step, d *StepDef) {
	f.scenarioFailed = true

	f.pretty.Undefined(s, d)
}

func (f *teamcityFormatter) Pending(s *gherkin.Step, d *StepDef) {
	f.scenarioFailed = true

	f.pretty.Pending(s, d)
}

func (f *teamcityFormatter) Summary() {
	f.checkAndCloseCurrentScenario()
	f.checkAndCloseCurrentFeature()

	f.pretty.Summary()
}

func (f *teamcityFormatter) checkAndCloseCurrentFeature() {
	if f.currentFeature != nil {
		line := fmt.Sprintf("##teamcity[testSuiteFinished name='Feature: %s']", f.escape(f.currentFeature.Name))
		
		f.printLine(line)
		f.currentFeature = nil
	}
}

func (f *teamcityFormatter) checkAndCloseCurrentScenario() {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	duration := now - f.scenarioStarted

	output := f.split.GetBufferAndReset()

	if f.currentScenario != nil {
		var line string
		if f.scenarioFailed {
			line = fmt.Sprintf("##teamcity[testFailed name='Scenario %s' details='%s']", f.escape(f.currentScenario.Name), f.escape(output))
			f.printLine(line)
		}
		
		line = fmt.Sprintf("##teamcity[testFinished name='Scenario %s' duration='%d']", f.escape(f.currentScenario.Name), duration)
		f.printLine(line)
		f.currentScenario = nil
	}
}

func (f *teamcityFormatter) printLine(line string) {
	f.out.Write([]byte(line + "\n"))
}

func (f *teamcityFormatter) escape(value string) string {
	buff := []rune{}

	for _, s := range stripansi.Strip(value) {
		switch s {
		case '\'':
			buff = append(buff, '|', '\'')
		case '\n':
			buff = append(buff, '|', 'n')
		case '\r':
			buff = append(buff, '|', 'r')
		case '|':
			buff = append(buff, '|', '|')
		case '[':
			buff = append(buff, '|', '[')
		case ']':
			buff = append(buff, '|', ']')
		default:
			buff = append(buff, s)
		}
	}

	return string(buff)
}

type splitWriter struct {
	out io.Writer
	buffer *bytes.Buffer
}

func newSplitWriter(writer io.Writer) *splitWriter {
	return &splitWriter{writer, bytes.NewBuffer([]byte{})}
}

func (w *splitWriter) Write(p []byte) (int, error) {
	n, err := w.out.Write(p)
	w.buffer.Write(p)

	return n, err
}

func (w *splitWriter) GetBufferAndReset() string {
	s := w.buffer.String()

	w.buffer.Reset()

	return s
}