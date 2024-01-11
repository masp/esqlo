package esqlo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bufHandler struct {
	resp []byte
}

func (b *bufHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(201)
	w.Write(b.resp)
}

func TestHandleRenderTemplate(t *testing.T) {
	db := &MemDB{
		Tables: map[string]*MemTable{
			"test_table": {
				Columns: []string{"name", "age"},
				Rows: [][]any{
					{"John", 20},
					{"Jane", 30},
				},
			},
		},
	}
	h := &bufHandler{resp: []byte(`<sql id="test">SELECT name FROM test_table</sql>
{{test[1].name}}
`)}
	dir := &Handler{
		Databases:  map[string]Database{ImplicitDb: db},
		fileserver: h,
	}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()
	dir.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	t.Logf("resp: %s", string(body))
	assert.Equal(t, string(body), "\nJane\n")
}
