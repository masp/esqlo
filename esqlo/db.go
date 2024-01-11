package esqlo

import (
	"fmt"
	"io"
	"net/url"

	"database/sql"

	_ "github.com/marcboeker/go-duckdb"
	"gopkg.in/yaml.v3"
)

const (
	ImplicitDb = "__mem__" // reserved scheme name for in-memory database. All tables are stored in this database for each page render.
)

type Database interface {
	OpenConnection(path *url.URL) error
	Query(query string) (*Result, error)
	Close() error
}

func LoadConfigs(r io.Reader) (map[string]Database, error) {
	var root map[string]yaml.Node
	err := yaml.NewDecoder(r).Decode(&root)
	if err != nil {
		return nil, err
	}

	configs := make(map[string]Database)
	for name, node := range root {
		cfg, err := loadCfgNode(node)
		if err != nil {
			return nil, err
		}
		configs[name] = cfg
	}
	return configs, nil
}

func loadCfgNode(node yaml.Node) (Database, error) {
	var typ struct {
		Field yaml.Node `yaml:"type"`
	}
	err := node.Decode(&typ)
	if err != nil {
		return nil, fmt.Errorf("[%d:%d] missing 'type' field: %w", node.Line, node.Column, err)
	}

	if typ.Field.Kind != yaml.ScalarNode {
		return nil, fmt.Errorf("[%d:%d] 'type' field must be a string", typ.Field.Line, typ.Field.Column)
	}

	switch typ.Field.Value {
	case "sqlite":
		db := &Sqlite{}
		err := node.Decode(db)
		return nil, err
	default:
		return nil, fmt.Errorf("[%d:%d] unknown database type: %s", typ.Field.Line, typ.Field.Column, typ.Field.Value)
	}
}

type Result struct {
	Columns []string
	Rows    []any // each row with a slice of values for each row. len(rows) == number of results, len(rows[0]) == len(ColumnNames)
}

type Sqlite struct {
	Path string `yaml:"path"`
}

func (s *Sqlite) OpenConnection(path *url.URL) error {
	return nil
}

type DuckDB struct {
	conn *sql.DB
}

func (d *DuckDB) OpenConnection(path *url.URL) (err error) {
	if path.Scheme == "duckdb" {
		path.Path = path.Host + path.Path
	} else {
		path.Path = ""
	}

	d.conn, err = sql.Open("duckdb", path.Path)
	return err
}

func (d *DuckDB) Query(query string) (*Result, error) {
	rows, err := d.conn.Query(query)
	if err == sql.ErrNoRows {
		cols, err := rows.Columns()
		return &Result{Columns: cols}, err
	} else if err != nil {
		return nil, err
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result, err := readRows(cols, rows)
	return &Result{Columns: cols, Rows: result}, err
}

func (d *DuckDB) Close() error {
	return d.conn.Close()
}

func readRows(cols []string, rows *sql.Rows) (result []any, err error) {
	for rows.Next() {
		var rowvals []any // ptr to any
		for i := 0; i < len(cols); i++ {
			var val any
			rowvals = append(rowvals, &val)
		}
		err := rows.Scan(rowvals...)
		if err != nil {
			return nil, err
		}

		values := make(map[string]any)
		for i, col := range cols {
			values[col] = *rowvals[i].(*any)
		}
		result = append(result, values)
	}
	return result, nil
}
