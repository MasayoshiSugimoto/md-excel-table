package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"path/filepath"

	"github.com/atotto/clipboard"
)

const MIN_COLUMN_WIDTH = 3
const USER_ONLY = 0600

func main() {
	fmt.Println("Start")

	tmpPath := filepath.Join(os.TempDir(), "md-table.txt")
	file, err := os.OpenFile(tmpPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, USER_ONLY)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	clip, err := clipboard.ReadAll()
	if err != nil {
		file.WriteString(fmt.Sprintf("Failed to read from clipboard: %v\n", err))
		log.Fatal(err)
	}
	file.WriteString(fmt.Sprintf("Read from clipboard: \n%s\n", clip))

	buf := new(bytes.Buffer)
	Convert(clip, buf)

	out := buf.String()

	file.WriteString(fmt.Sprintf("Writing to clipboard: \n%s\n", out))

	err = clipboard.WriteAll(out)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done")
}

// Convert a markdown table from one format to the other
func Convert(table string, out io.Writer) {
	// Detect table format
	lines := strings.Split(strings.ReplaceAll(table, "\r", ""), "\n")
	if len(lines) == 0 {
		return
	}

	// Remove empty lines
	if len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	if strings.Count(table, `|`) > strings.Count(table, "\t") { // Convert from markdown to tsv
		FromMdToTsv(lines, out)
	} else { // Convert from tsv to markdow
		ToMarkDown(ParseExcelTable(lines), out)
	}
}

func FromMdToTsv(lines []string, out io.Writer) {
	// Remove '|' at the start and at the end
	startRe := regexp.MustCompile(`^\|`)
	endRe := regexp.MustCompile(`\|$`)
	for i, line := range lines {
		lines[i] = endRe.ReplaceAllString(startRe.ReplaceAllString(line, ""), "")
	}

	if len(lines) == 0 {
		return
	}

	first := true
	for _, line := range lines {
		if !first {
			io.WriteString(out, "\n")
		}
		first = false
		row := strings.Split(line, `|`)
		for j, cell := range row {
			row[j] = strings.Trim(cell, " \t")
		}
		io.WriteString(out, strings.Join(row, "\t"))
	}
}

/******************************************************************************
* MdTable
******************************************************************************/

type MdTable struct {
	width        int
	header       []string
	data         [][]string
	columnsWidth []int
	alignments   []Alignment
}

type Alignment int

const (
	AlignNone = iota
	AlignLeft
	AlignCenter
	AlignRight
)

/*
Parse tab separated string comming from Excel and convert it to a data structure.

If the string has been pasted from markdown once to Excel, the second line might
contain a header separator.
*/
func ParseExcelTable(lines []string) *MdTable {
	const HeaderIndex = 0
	const HeaderSeparatorIndex = 1

	table := make([][]string, len(lines))

	// Separate values by tab
	for i, line := range lines {
		table[i] = strings.Split(line, "\t")
	}

	// Cleanup trailing tabs
	header := table[HeaderIndex]
	if len(header[len(header)-1]) == 0 {
		for i, row := range table {
			table[i] = row[:len(header)-1]
		}
	}

	// Check if we have header separator
	isHeaderSeparator := false
	if len(table) >= 2 {
		isHeaderSeparator = true
		for _, cell := range table[HeaderSeparatorIndex] {
			if strings.Count(cell, "-") < 3 {
				isHeaderSeparator = false
			}
		}
	}

	// Calculate columns width
	columnsWidth := make([]int, len(table[HeaderIndex]))
	for i, row := range table {
		for j, cell := range row {
			length := len(cell)
			if isHeaderSeparator && i == HeaderSeparatorIndex {
				length = MIN_COLUMN_WIDTH
			}
			if length > columnsWidth[j] {
				columnsWidth[j] = length
			}
		}
	}

	// Calculate alignments
	alignments := make([]Alignment, len(table[HeaderIndex]))
	for i := 0; i < len(table[HeaderIndex]); i++ {
		alignments[i] = AlignNone
	}
	if isHeaderSeparator {
		centeredRe := regexp.MustCompile(`:-+:`)
		leftAlignRe := regexp.MustCompile(`:-+`)
		rightAlignRe := regexp.MustCompile(`-+:`)
		for i, cell := range table[HeaderSeparatorIndex] {
			if centeredRe.MatchString(cell) {
				alignments[i] = AlignCenter
			} else if leftAlignRe.MatchString(cell) {
				alignments[i] = AlignLeft
			} else if rightAlignRe.MatchString(cell) {
				alignments[i] = AlignRight
			} else {
				alignments[i] = AlignNone
			}
		}
	}

	dataStartIndex := 1
	if isHeaderSeparator {
		dataStartIndex = 2
	}

	return &MdTable{
		len(table[HeaderIndex]),
		table[HeaderIndex],
		table[dataStartIndex:len(table)],
		columnsWidth,
		alignments,
	}
}

func ToMarkDown(mdTable *MdTable, out io.Writer) {

	// Print the header
	fmt.Fprint(out, mdTablePrintRow(mdTable, mdTable.header)+"\n")

	{ // Print header separator
		row := make([]string, mdTable.width)
		for i := 0; i < mdTable.width; i++ {
			var dashes []byte
			for x := 0; x < mdTable.columnsWidth[i]; x++ {
				dashes = append(dashes, '-')
			}
			switch mdTable.alignments[i] {
			case AlignNone:
				row[i] = fmt.Sprintf("-%s-", string(dashes))
			case AlignLeft:
				row[i] = fmt.Sprintf(":%s-", string(dashes))
			case AlignCenter:
				row[i] = fmt.Sprintf(":%s:", string(dashes))
			case AlignRight:
				row[i] = fmt.Sprintf("-%s:", string(dashes))
			default:
				row[i] = fmt.Sprintf("-%s-", string(dashes))
			}
		}

		if len(mdTable.data) > 0 {
			fmt.Fprintf(out, "|%s|\n", strings.Join(row, "|"))
		} else {
			fmt.Fprintf(out, "|%s|", strings.Join(row, "|"))
		}
	}

	// Print data
	for i, row := range mdTable.data {
		if i+1 >= len(mdTable.data) {
			fmt.Fprint(out, mdTablePrintRow(mdTable, row))
		} else {
			fmt.Fprint(out, mdTablePrintRow(mdTable, row)+"\n")
		}
	}

}

func padLeft(s string, length int) string {
	var b bytes.Buffer
	b.WriteString(s)
	for i := 0; i < length-len(s); i++ {
		b.WriteString(" ")
	}
	return b.String()
}

func padRight(s string, length int) string {
	var b bytes.Buffer
	for i := 0; i < length-len(s); i++ {
		b.WriteString(" ")
	}
	b.WriteString(s)
	return b.String()
}

func padCenter(s string, length int) string {
	spaceCount := length - len(s)
	var leftSpaces = spaceCount / 2
	if spaceCount%2 != 0 {
		leftSpaces = (spaceCount / 2) + 1
	}
	var b bytes.Buffer
	for i := 0; i < leftSpaces; i++ {
		b.WriteString(" ")
	}
	b.WriteString(s)
	for i := 0; i < length-len(s)-leftSpaces; i++ {
		b.WriteString(" ")
	}
	return b.String()
}

func mdTablePrintRow(mdTable *MdTable, row []string) string {
	r := make([]string, len(row))
	for j, cell := range row {
		switch mdTable.alignments[j] {
		case AlignNone:
			r[j] = padLeft(cell, mdTable.columnsWidth[j])
		case AlignLeft:
			r[j] = padLeft(cell, mdTable.columnsWidth[j])
		case AlignCenter:
			r[j] = padCenter(cell, mdTable.columnsWidth[j])
		case AlignRight:
			r[j] = padRight(cell, mdTable.columnsWidth[j])
		default:
			r[j] = padLeft(cell, mdTable.columnsWidth[j])
		}
	}
	return fmt.Sprintf("| %s |", strings.Join(r, " | "))
}
