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

		/******************************************************************************
		 * Markdown to TSV
		 ******************************************************************************/

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
		// Empty cell
		{
			`| Tables   |      Are      |  Cool |
|----------|:-------------:|------:|
| col 1 is |               | $1600 |`,
			`Tables	Are	Cool
----------	:-------------:	------:
col 1 is		$1600`,
		},
		// No space arount `|`
		{
			`|Tables|Are|Cool|
|---|---|---|
|col 1 is||$1600|`,
			`Tables	Are	Cool
---	---	---
col 1 is		$1600`,
		},
		// Just header
		{
			`|Tables|Are|Cool|`,
			`Tables	Are	Cool`,
		},

		/******************************************************************************
		 * TSV to markdown
		 ******************************************************************************/

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
		// Ignores trailing tabs
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
		// No record, no header separator
		{
			`Tables	Are	Cool`,
			`| Tables | Are | Cool |
|--------|-----|------|`,
		},
		// No record, with header separator
		{
			`Tables	Are	Cool	
--------	:-----:	------:`,
			`| Tables | Are | Cool |
|--------|:---:|-----:|`,
		},
		// Unaligned header
		{
			`Tables	Are	Cool
---	:---:	---:
col 1 is	left-aligned	$1600
col 2 is	centered	$12
col 3 is	right-aligned	$1`,
			`| Tables   |      Are      |  Cool |
|----------|:-------------:|------:|
| col 1 is |  left-aligned | $1600 |
| col 2 is |    centered   |   $12 |
| col 3 is | right-aligned |    $1 |`,
		},
		// Almost header
		{
			`Tables	Are	Cool
---	--	---
col 1 is	left-aligned	$1600
col 2 is	centered	$12
col 3 is	right-aligned	$1`,
			`| Tables   | Are           | Cool  |
|----------|---------------|-------|
| ---      | --            | ---   |
| col 1 is | left-aligned  | $1600 |
| col 2 is | centered      | $12   |
| col 3 is | right-aligned | $1    |`,
		},
		// Empty cells
		{
			`x	y	z
:---:	:---:	:---:
	1	123456`,
			`|  x  |  y  |    z   |
|:---:|:---:|:------:|
|     |  1  | 123456 |`,
		},
		// Extra trailing tabs
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
