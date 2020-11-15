package main

import (
	"bytes"
	"testing"
)

func TestConvert(t *testing.T) {

	var scenarios = []struct {
		in  string
		out string
	}{
		// Markdown to TSV
		{
			`| Tables   |      Are      |  Cool |
|----------|:-------------:|------:|
| col 1 is |  left-aligned | $1600 |
| col 2 is |    centered   |   $12 |
| col 3 is | right-aligned |    $1 |`,
			`Tables	Are	Cool
----------	:-------------:	------:
col 1 is	left-aligned	$1600
col 2 is	centered	$12
col 3 is	right-aligned	$1`,
		},
		// TSV to markdown
		{
			`Tables	Are	Cool
----------	:-------------:	------:
col 1 is	left-aligned	$1600
col 2 is	centered	$12
col 3 is	right-aligned	$1`,
			`| Tables   |      Are      |  Cool |
|----------|:-------------:|------:|
| col 1 is |  left-aligned | $1600 |
| col 2 is |    centered   |   $12 |
| col 3 is | right-aligned |    $1 |`,
		},
		// No header delimiter in TSV
		{
			`Tables	Are	Cool
col 1 is	left-aligned	$1600
col 2 is	centered	$12
col 3 is	right-aligned	$1`,
			`| Tables   | Are           | Cool  |
|----------|---------------|-------|
| col 1 is | left-aligned  | $1600 |
| col 2 is | centered      | $12   |
| col 3 is | right-aligned | $1    |`,
		},
	}

	for _, test := range scenarios {
		buf := new(bytes.Buffer)
		Convert(test.in, buf)
		if test.out != buf.String() {
			t.Errorf("Got:\n%s\nInstead of:\n%s", buf.String(), test.out)
		}
	}
}
