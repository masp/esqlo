package esqlo

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func TestLineCounterHTML(t *testing.T) {

	// simple test case against HTML, let's see if it works when we put all this together

	h := `<!doctype html>

<html>
	<body>
        <div id="testing">blah</div>
	</body>
</html>`

	hb := []byte(h)

	lc := NewLineCounter(bytes.NewReader(hb))
	z := html.NewTokenizer(lc)
	offset := 0
	divOffset := -1
	var divRaw string
	for {
		tt := z.Next()
		p := offset
		offset += len(z.Raw())
		if tt == html.ErrorToken {
			break
		}
		if tt == html.StartTagToken && z.Token().Data == "div" {
			divOffset = p
			divRaw = string(z.Raw())
		}
	}

	line, col := lc.LineCol(divOffset)
	assert.Equal(t, `<div id="testing">`, divRaw)
	assert.Equal(t, 5, line)
	assert.Equal(t, 9, col)
}
