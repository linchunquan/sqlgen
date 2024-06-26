package schema

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"text/tabwriter"
)

type base struct {
	Dialect Dialect
}

// Table returns a SQL statement to create the table.
func (b *base) Table(t *Table) string {

	// use a large default buffer size of so that
	// the tabbing doesn't get prematurely flushed
	// resulting in un-even lines.
	var byt = make([]byte, 0, 100000)
	var buf = bytes.NewBuffer(byt)

	// use a tab writer to evenly space the column
	// names and column types.
	var tab = tabwriter.NewWriter(buf, 0, 8, 1, ' ', 0)
	b.columnw(tab, t, t.Fields, false, false, true)

	// flush the tab writer to write to the buffer
	tab.Flush()

	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s\n);", t.Name, buf.String())
}

// Index returns a SQL statement to create the index.
func (b *base) Index(table *Table, index *Index) string {
	log.Printf("create index:%+v", index)
	var obj = "INDEX"
	if index.Unique {
		obj = "UNIQUE INDEX"
	}
	return fmt.Sprintf("CREATE %s IF NOT EXISTS %s ON %s (%s)", obj, index.Name, table.Name, b.columns(nil, index.Fields, true, false, false))
}

// Foreign returns a SQL statement to add foreign key.
func (b *base)Foreign(table *Table, foreign *Foreign) string {
	log.Printf("create foreign key:%+v", foreign)
	fromColumns := strings.Join(foreign.FromColumns, ",")
	toColumns := strings.Join(foreign.ToColumns, ",")
	return fmt.Sprintf("ALTER TABLE %s ADD FOREIGN KEY (%s) REFERENCES %s (%s);", table.Name, fromColumns, foreign.ToTable, toColumns)
}

func (b *base) Insert(t *Table) string {
	var fields []*Field
	var params []string
	var i int

	for _, field := range t.Fields {
		if !field.Auto {
			fields = append(fields, field)
			params = append(params, b.Dialect.Param(i))
			i++
		}
	}

	return fmt.Sprintf("INSERT INTO %s (%s\n) VALUES (%s)", t.Name, b.columns(nil, fields, false, false, false), strings.Join(params, ","))
}

func (b *base) Update(t *Table, fields []*Field) string {
	return fmt.Sprintf("UPDATE %s SET %s %s", t.Name, b.columns(nil, t.Fields, false, true, false), b.clause(fields, len(t.Fields)))
}

func (b *base) Delete(t *Table, fields []*Field) string {
	return fmt.Sprintf("DELETE FROM %s %s", t.Name, b.clause(fields, 0))
}

func (b *base) Select(t *Table, fields []*Field) string {
	return fmt.Sprintf("SELECT %s\nFROM %s %s", b.columns(t, t.Fields, false, false, false), t.Name, b.clause(fields, 0))
}

func (b *base) SelectRange(t *Table, fields []*Field) string {
	return fmt.Sprintf("SELECT %s\nFROM %s %s\nLIMIT %s OFFSET %s", b.columns(t, t.Fields, false, false, false), t.Name, b.clause(fields, 0), b.Dialect.Param(len(fields)), b.Dialect.Param(len(fields)+1))
}

func (b *base) SelectCount(t *Table, fields []*Field) string {
	return fmt.Sprintf("SELECT count(1)\nFROM %s %s", t.Name, b.clause(fields, 0))
}

func (b *base) SelectByUniqueIndex(t *Table, fields []*Field, index *Index) string{
	return ""
}

// Param returns the parameters symbol used in prepared
// sql statements.
func (b *base) Param(i int) string {
	return "?"
}

// Column returns a SQL type for the given field.
//
// For Mysql and Postgres see:
// https://github.com/eaigner/hood/blob/master/mysql.go#L35
func (b *base) Column(f *Field) string {
	switch f.Type {
	case INTEGER:
		return "INTEGER"
	case BOOLEAN:
		return "BOOLEAN"
	case BLOB:
		return "BLOB"
	case VARCHAR:
		return "TEXT"
	default:
		return "TEXT"
	}
}

// Token returns the SQL string for the requested token.
func (b *base) Token(v int) (_ string) {
	switch v {
	case AUTO_INCREMENT:
		return "AUTOINCREMENT"
	case PRIMARY_KEY:
		return "PRIMARY KEY"
	default:
		return
	}
}

// helper function to generate a block of columns. You
// can optionally generate in inline list of columns,
// include an assignment operator, and include column
// definitions.
func (b *base) columns(table *Table, fields []*Field, inline, assign, def bool) string {
	var buf bytes.Buffer
	b.columnw(&buf, table, fields, inline, assign, def)
	return buf.String()
}

// helper function to write a block of columns to w.
func (b *base) columnw(w io.Writer, table *Table, fields []*Field, inline, assign, def bool) {

	for i, field := range fields {
		if !inline {
			io.WriteString(w, "\n")
		}

		switch {
		case i == 0 && !inline:
			io.WriteString(w, " ")
		case i != 0:
			io.WriteString(w, ",")
		}
		if table!=nil{
			//io.WriteString(w, table.Name+"."+field.Name)
			// 带 table.Name+"." 之后，在信创数据库无法识别
			io.WriteString(w, field.Name)
		}else{
			io.WriteString(w, field.Name)
		}


		if assign {
			io.WriteString(w, "=")
			io.WriteString(w, b.Dialect.Param(i))
		}

		if !def {
			continue
		}

		io.WriteString(w, "\t")
		io.WriteString(w, b.Dialect.Column(field))

		if field.Primary {
			io.WriteString(w, " ")
			io.WriteString(w, b.Dialect.Token(PRIMARY_KEY))
		}

		if field.Auto {
			io.WriteString(w, " ")
			io.WriteString(w, b.Dialect.Token(AUTO_INCREMENT))
		}
	}
}

// helper function to generate the Where clause
// section of a SQL statement
func (b *base) clause(fields []*Field, pos int) string {
	var buf bytes.Buffer

	var i int
	for _, field := range fields {
		buf.WriteString("\n")
		switch {
		case i == 0:
			buf.WriteString("WHERE")
		default:
			buf.WriteString("AND")
		}

		buf.WriteString(" ")
		if len(field.Operator)==0||strings.EqualFold("=",field.Operator){
			buf.WriteString(field.Name)
			buf.WriteString("=")
			buf.WriteString(b.Dialect.Param(i + pos))
		}else if strings.EqualFold(strings.ToUpper(field.Operator), "LIKE"){
			buf.WriteString(field.Name)
			buf.WriteString(" LIKE ")
			buf.WriteString(b.Dialect.Param(i + pos))
		}else{
			buf.WriteString(field.Operator)
			buf.WriteString("(")
			if field.ValueAsFirstArg{
				buf.WriteString(b.Dialect.Param(i + pos))
				buf.WriteString(",")
				buf.WriteString(field.Name)
			}else{
				buf.WriteString(field.Name)
				buf.WriteString(",")
				buf.WriteString(b.Dialect.Param(i + pos))
			}
			buf.WriteString(")")
		}

		i++
	}
	return buf.String()
}
