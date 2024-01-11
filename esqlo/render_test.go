package esqlo

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDb = &MemDB{
	Tables: map[string]*MemTable{
		"persons": {
			Columns: []string{"name", "age"},
			Rows: [][]any{
				{"John", 20},
				{"Jane", 30},
			},
		},
	},
}

func TestStripSqlTags(t *testing.T) {
	renderer := NewRenderer()
	renderer.Databases[ImplicitDb] = testDb

	var out bytes.Buffer
	src := `<html><body><sql ignore id="p">SELECT * FROM persons</sql><p class={{p.age}}>hello</p></body></html>`
	err := renderer.RenderHTML(strings.NewReader(src), &out)
	require.NoError(t, err)
	assert.Equal(t, `<html><body><p class=20>hello</p></body></html>`, out.String())
	t.Logf("context: %+v", renderer.context)
	assert.Len(t, renderer.Errors, 0)
}

func TestRenderSection(t *testing.T) {
	renderer := NewRenderer()
	renderer.Databases[ImplicitDb] = testDb

	var out bytes.Buffer
	src := `
<sql id="p">SELECT * FROM persons</sql>
{{#p}}
<p>{{name}} is {{age}} years old</p>
{{/p}}`
	err := renderer.RenderHTML(strings.NewReader(src), &out)
	require.NoError(t, err)
	assert.Equal(t, `

<p>John is 20 years old</p>
<p>Jane is 30 years old</p>
`, out.String())
	t.Logf("context: %+v", renderer.context)
	assert.Len(t, renderer.Errors, 0)
}
