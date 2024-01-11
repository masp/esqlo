package esqlo

import (
	"fmt"
	"net/url"

	"github.com/xwb1989/sqlparser"
)

type MemDB struct {
	Tables map[string]*MemTable
}

type MemTable struct {
	Columns []string
	Rows    [][]any
}

func (db *MemDB) OpenConnection(path *url.URL) error {
	return nil
}

func (db *MemDB) Query(query string) (*Result, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return nil, err
	}

	sel, ok := stmt.(*sqlparser.Select)
	if !ok {
		return nil, fmt.Errorf("only SELECT statements are supported")
	}

	if len(sel.From) != 1 {
		return nil, fmt.Errorf("only single table SELECT statements are supported")
	}

	tableName := sqlparser.String(sel.From[0])
	table, ok := db.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("table %q not found", tableName)
	}

	var columns []string
	for _, expr := range sel.SelectExprs {
		switch expr := expr.(type) {
		case *sqlparser.AliasedExpr:
			columns = append(columns, sqlparser.String(expr.Expr))
		case *sqlparser.StarExpr:
			columns = append(columns, table.Columns...)
		default:
			return nil, fmt.Errorf("only bare column names (no expressions) are supported")
		}
	}

	var indices []int
	for _, col := range columns {
		found := false
		for i, name := range table.Columns {
			if name == col {
				indices = append(indices, i)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("column %q not found", col)
		}
	}

	var rows []any
	for _, row := range table.Rows {
		values := make(map[string]any)
		for _, i := range indices {
			values[table.Columns[i]] = row[i]
		}
		rows = append(rows, values)
	}

	result := &Result{
		Columns: columns,
		Rows:    rows,
	}
	return result, nil
}

func (db *MemDB) Close() error {
	return nil
}
