package esqlo

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// func assertCast[T any](t *testing.T, v T, typ any) T {
// 	t.Helper()
// 	c, ok := typ.(T)
// 	assert.True(t, ok, "type assertion failed: %T is not %T", typ, v)
// 	return c
// }

// func TestLoadConifg(t *testing.T) {
// 	doc, err := os.Open("testdata/sample_config.yml")
// 	require.NoError(t, err)
// 	configs, err := LoadConfigs(doc)
// 	if err != nil {
// 		t.Fatalf("load config: %v", err)
// 	}

// 	assert.Equal(t, 2, len(configs))
// 	var cfg *Sqlite
// 	cfg = assertCast(t, cfg, configs["persons"])
// 	assert.Equal(t, "./persons.db", cfg.Path)
// 	cfg = assertCast(t, cfg, configs["famous_persons"])
// 	assert.Equal(t, "./famous_persons.db", cfg.Path)
// }

func parseUrl(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func TestReadDuckDb(t *testing.T) {
	db := &DuckDB{}
	err := db.OpenConnection(parseUrl("duckdb://./testdata/people.duckdb"))
	require.NoError(t, err)
	defer db.Close()

	res, err := db.Query("SELECT id, name FROM people")
	require.NoError(t, err)
	assert.Equal(t, []string{"id", "name"}, res.Columns)
	assert.Equal(t, []any{
		map[string]any{"name": "John", "id": int32(42)},
		map[string]any{"name": "Jane", "id": int32(41)},
	}, res.Rows)
}

func TestReadCsv(t *testing.T) {
	db := &DuckDB{}
	err := db.OpenConnection(parseUrl("duckdb"))
	require.NoError(t, err)
	defer db.Close()

	res, err := db.Query("SELECT id, name FROM 'testdata/people.csv'")
	require.NoError(t, err)
	assert.Equal(t, []string{"id", "name"}, res.Columns)
	assert.Equal(t, []any{
		map[string]any{"name": "John", "id": int64(42)},
		map[string]any{"name": "Jane", "id": int64(41)},
	}, res.Rows)
}
