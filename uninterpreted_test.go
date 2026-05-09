package ishell

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// linesReader returns a function that yields the supplied lines one at a
// time, then returns io.EOF on subsequent calls.
func linesReader(lines []string) func() (string, error) {
	i := 0
	return func() (string, error) {
		if i >= len(lines) {
			return "", io.EOF
		}
		l := lines[i]
		i++
		return l, nil
	}
}

func TestReadMultiLinesAccumFromReaderTerminator(t *testing.T) {
	// Predicate stops accumulation once the buffer contains "GO", regardless of
	// per-line semicolon terminators.
	read := linesReader([]string{"select 1;", "select 2;", "GO"})
	multiCalls := []bool{}
	pred := func(line, accumulated string) bool {
		return !strings.Contains(accumulated, "GO")
	}
	out, err := readMultiLinesAccumFromReader(read, func(b bool) { multiCalls = append(multiCalls, b) }, pred)
	assert.NoError(t, err)
	assert.Equal(t, "select 1;\nselect 2;\nGO", out)
	// setMulti(true) once at currentLine==1, setMulti(false) once at end.
	assert.Equal(t, []bool{true, false}, multiCalls)
}

func TestReadMultiLinesAccumFromReaderSingleLine(t *testing.T) {
	// First line ends accumulation; setMulti is never called.
	read := linesReader([]string{"select 1;"})
	multiCalls := []bool{}
	pred := func(line, accumulated string) bool { return false }
	out, err := readMultiLinesAccumFromReader(read, func(b bool) { multiCalls = append(multiCalls, b) }, pred)
	assert.NoError(t, err)
	assert.Equal(t, "select 1;", out)
	assert.Empty(t, multiCalls)
}

func TestReadMultiLinesAccumFromReaderEmpty(t *testing.T) {
	// First line is empty and predicate returns false: shell never enters
	// multi-mode; output is empty. This mirrors the empty-Enter path.
	read := linesReader([]string{""})
	multiCalls := []bool{}
	pred := func(line, accumulated string) bool { return false }
	out, err := readMultiLinesAccumFromReader(read, func(b bool) { multiCalls = append(multiCalls, b) }, pred)
	assert.NoError(t, err)
	assert.Equal(t, "", out)
	assert.Empty(t, multiCalls)
}

func TestReadMultiLinesAccumFromReaderEOFMidStream(t *testing.T) {
	// EOF arriving after the predicate has chosen to keep reading surfaces
	// the EOF error and returns the buffer collected so far. The trailing
	// "\n" is the inter-line separator written by fmt.Fprintln before the
	// next readLine call returned EOF.
	read := linesReader([]string{"select 1;"})
	pred := func(line, accumulated string) bool { return true }
	out, err := readMultiLinesAccumFromReader(read, func(b bool) {}, pred)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, "select 1;\n", out)
}

