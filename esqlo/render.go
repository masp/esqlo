package esqlo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/masp/esqlo/esqlo/mustache"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

type SqlTag struct {
	Offset, End int      // the offset in the source file where this tag starts
	Src         string   // the database connection to use
	Database    Database // the database connection to use
	TableName   string   // the name to store this table as

	Query  string  // the sql query to execute
	Result *Result // the result of the query
}

type Err struct {
	Line, Col int
	Msg       error
}

func (e *Err) Error() string {
	return fmt.Sprintf("[%d:%d] %s", e.Line, e.Col, e.Msg)
}

type Renderer struct {
	Databases map[string]Database
	Errors    []*Err // any errors that occurred while processing the document (can be ignored gracefully)

	lc           *LineCounter
	allSqlTags   []*SqlTag // all sql tags in the document, irregardless of scope, in order of appearance
	activeSqlTag *SqlTag   // the sql tag that is currently being tokenized (nil if not in one)

	context map[string]any // the context to use when rendering mustache tags
}

func NewRenderer() *Renderer {
	return &Renderer{
		Databases: make(map[string]Database),
		context:   make(map[string]any),
	}
}

func (r *Renderer) errorf(offset int, format string, args ...interface{}) {
	l, c := r.lc.LineCol(offset)
	r.Errors = append(r.Errors, &Err{
		Line: l,
		Col:  c,
		Msg:  fmt.Errorf(format, args...),
	})
}

func (r *Renderer) errlist() error {
	if len(r.Errors) == 0 {
		return nil
	}
	errs := make([]error, len(r.Errors))
	for i := 0; i < len(r.Errors); i++ {
		errs[i] = r.Errors[i]
	}
	return errors.Join(errs...)
}

// / RenderHTML renders an HTML document with embedded SQL statements. The SQL statements are resolved using the provided
// / database connections. The rendered HTML is written to the provided writer.
//
// The SQL statements are embedded in the HTML using the following syntax:
//
//		<!-- Define a table for this HTML document that can be used in later in the document to inject live data. -->
//		<!-- The src must be a defined database with name persons, and the query must be a valid SQL query for that database type. -->
//		<sql src="sqlite://persons.sqlite" name="famous_persons">SELECT name, age FROM persons WHERE is_famous={{query.is_famous}} LIMIT {{query.page_size}}"</sql>
//
//		<p>Here is a list of famous people:</p>
//	 	<p>{{persons.name[0]}} is {{persons.age[0]}} years old</p>
//	 	<p>{{persons.name[1]}} is {{persons.age[1]}} years old</p>
//	 	<p>{{persons.name[2]}} is {{persons.age[2]}} years old</p>
//
// RenderHTML does not modify the tree, besides removing the <sql> tags at the start of the html document.
//
// Replacement values are injected into the HTML document afterwards using the Mustache template syntax.
func (r *Renderer) RenderHTML(src io.Reader, w io.Writer) error {
	var buf bytes.Buffer
	r.walkTokens(src, &buf)
	r.renderMustache(buf.String(), w)
	return r.errlist()
}

func (r *Renderer) walkTokens(src io.Reader, w io.Writer) {
	r.lc = NewLineCounter(src)
	z := html.NewTokenizer(r.lc)
	var offset int
	for {
		tt := z.Next()
		p := offset
		offset += len(z.Raw())
		switch tt {
		case html.ErrorToken:
			err := z.Err()
			if err == io.EOF {
				err = nil
				return
			} else if err != nil {
				r.errorf(p, err.Error())
			}
			r.render(w, z.Raw())
		case html.StartTagToken:
			tn, hasAttr := z.TagName()
			if bytes.Equal(tn, []byte("sql")) {
				if r.activeSqlTag != nil {
					r.errorf(p, "nested sql tags are not allowed")
				}

				r.activeSqlTag = &SqlTag{Offset: offset}
				if hasAttr {
					for {
						k, v, more := z.TagAttr()
						if bytes.Equal(k, []byte("src")) {
							r.activeSqlTag.Src = string(v)
						} else if bytes.Equal(k, []byte("id")) {
							r.activeSqlTag.TableName = string(v)
						}
						if !more {
							break
						}
					}
				}

				tag := r.activeSqlTag
				if tag.Src == "" {
					tag.Src = ImplicitDb // implicit schema for in memory
				}

				srcUrl, err := url.Parse(tag.Src)
				if err != nil {
					r.errorf(p, "invalid src attribute: %v", err)
					continue
				}

				tag.Database, err = r.LoadDatabase(srcUrl)
				if err != nil {
					r.errorf(p, "loading database: %w", err)
					continue
				}

				if r.activeSqlTag.TableName == "" {
					r.errorf(p, "missing required 'id' attribute")
				}
			} else {
				r.render(w, z.Raw())
			}
		case html.TextToken:
			if r.activeSqlTag != nil {
				r.activeSqlTag.Query += string(z.Raw())
			} else {
				r.render(w, z.Raw())
			}
		case html.SelfClosingTagToken:
			if z.Token().Data == "sql" {
				continue // ignore
			} else {
				r.render(w, z.Raw())
			}
		case html.EndTagToken:
			if z.Token().Data == "sql" {
				if r.activeSqlTag == nil {
					r.errorf(offset, "unexpected end tag </sql>")
					continue
				}
				r.activeSqlTag.End = offset
				r.loadSql(r.activeSqlTag)
				r.activeSqlTag = nil
			} else {
				r.render(w, z.Raw())
			}
		case html.CommentToken, html.DoctypeToken:
			if r.activeSqlTag == nil {
				r.render(w, z.Raw())
			}
		}
	}
}

func (r *Renderer) render(w io.Writer, src []byte) {
	w.Write(src)
}

func (r *Renderer) renderMustache(src string, w io.Writer) {
	w.Write([]byte(mustache.Render(src, r.context)))
}

func (r *Renderer) LoadDatabase(path *url.URL) (Database, error) {
	if path.Path == ImplicitDb {
		return r.Databases[ImplicitDb], nil
	}

	if path.Path == "duckdb" || path.Scheme == "duckdb" {
		db := &DuckDB{}
		return db, db.OpenConnection(path)
	}
	return nil, fmt.Errorf("unknown database: %s", path)
}

func (r *Renderer) loadSql(tag *SqlTag) {
	var err error
	tag.Result, err = tag.Database.Query(tag.Query)
	if err != nil {
		r.errorf(tag.Offset, "executing query: %v", err)
		return
	}

	if tag.TableName == "" {
		return // error already recorded at start tag
	}
	log.Debug().Msgf("loaded table %q with %d rows (columns: %+v)", tag.TableName, len(tag.Result.Rows), tag.Result.Columns)
	r.context[tag.TableName] = tag.Result.Rows
	r.allSqlTags = append(r.allSqlTags, r.activeSqlTag)
}
