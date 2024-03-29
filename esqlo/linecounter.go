package esqlo

import (
	"io"
	"sort"
)

// LineCounter wraps an io.Reader and records line information to allow converting offsets into line numbers.
// Line feed (ASCII 10, '\n') is used as the line ending (files using CRLF line termination will of course
// count correctly as well).
type LineCounter struct {
	r        io.Reader
	lineoffs []int // offset of each line ending
	offset   int   // current offset after last read
}

// NewLineCounter returns a new LineCounter reading from r.
func NewLineCounter(r io.Reader) *LineCounter {
	return &LineCounter{r: r}
}

// Read calls through the underlying Reader and examines the returned data for line information
// before returning it.
func (l *LineCounter) Read(p []byte) (n int, err error) {
	n, err = l.r.Read(p)
	l.scan(p[:n])
	return
}

// LineCol returns, give a byte offset of the input, the
// line number and that line's byte offset in the input.
// Line numbers are 1-based, i.e. passing 0 returns line number 1.
// Offsets beyond the last recorded line will return the last recorded line.
func (l *LineCounter) LineCol(offset int) (line, col int) {

	i := sort.SearchInts(l.lineoffs, offset)

	var lineOffset int
	if i == 0 {
		lineOffset = 0 // special case for first line, has no prior line ending
	} else {
		lineOffset = l.lineoffs[i-1] + 1
	}

	line = i + 1
	col = offset - lineOffset + 1
	return
}

func (l *LineCounter) scan(p []byte) {
	for i := 0; i < len(p); i++ {
		if p[i] == '\n' {
			l.lineoffs = append(l.lineoffs, l.offset+i)
		}
	}
	l.offset += len(p)
}
