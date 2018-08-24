package main

import (
	"fmt"
	"io"
	"strings"

	"bitbucket.org/pkg/inflect"
	"github.com/linchunquan/sqlgen/schema"
	"path/filepath"
	"os"
	"bytes"
	"log"
)

func isPathExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
// writeSchema writes SQL statements to CREATE, INSERT,
// UPDATE and DELETE values from Table t.
func writeSchema(w io.Writer, d schema.Dialect, t *schema.Table, outputSqlFilePath string) {

	var outputSql = len(outputSqlFilePath) > 0
	var err error
	var sqlFile *os.File
	var sqlFileContent = &bytes.Buffer{}
	if outputSql{
		basepath:=filepath.Dir(outputSqlFilePath)
		if !isPathExist(basepath){
			if os.MkdirAll(basepath, os.ModePerm) != nil {
				panic("Unable to create directory:"+basepath+" for tagfile!")
			}
		}
		sqlFile, err = os.OpenFile(outputSqlFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err!=nil{
			panic(fmt.Errorf("create sql output file error:%v",err))
		}
	}

	writeConst(sqlFileContent, w,
		d.Table(t),
		"create", inflect.Singularize(t.Name), "stmt",
	)

	writeConst(nil, w,
		d.Insert(t),
		"insert", inflect.Singularize(t.Name), "stmt",
	)

	writeConst(nil, w,
		d.Select(t, nil),
		"select", inflect.Singularize(t.Name), "stmt",
	)

	writeConst(nil, w,
		d.SelectRange(t, nil),
		"select", inflect.Singularize(t.Name), "range", "stmt",
	)

	writeConst(nil, w,
		d.SelectCount(t, nil),
		"select", inflect.Singularize(t.Name), "count", "stmt",
	)

	if len(t.Primary) != 0 {
		writeConst(nil, w,
			d.Select(t, t.Primary), "select", inflect.Singularize(t.Name), "by", joinField(t.Primary, "And"), "stmt",
		)

		writeConst(nil, w,
			d.Update(t, t.Primary), "update", inflect.Singularize(t.Name), "by", joinField(t.Primary, "And"), "stmt",
		)

		writeConst(nil, w,
			d.Delete(t, t.Primary), "delete", inflect.Singularize(t.Name), "by", joinField(t.Primary, "And"), "stmt",
		)
	}

	for _, ix := range t.Index {

		writeConst(sqlFileContent, w,
			d.Index(t, ix),
			"create", ix.Name, "stmt",
		)

		writeConst(nil, w,
			d.Select(t, ix.Fields),
			"select", inflect.Singularize(t.Name), "by", joinField(ix.Fields, "And"), "stmt",
		)

		writeConst(nil, w,
			d.SelectCount(t, ix.Fields),
			"select", inflect.Singularize(t.Name), "count", "by", joinField(ix.Fields, "And"), "stmt",
		)

		if !ix.Unique {
			writeConst(nil, w,
				d.SelectRange(t, ix.Fields),
				"select", inflect.Singularize(t.Name), "range", "by", joinField(ix.Fields, "And"), "stmt",
			)
		} else {
			writeConst(nil, w,
				d.Update(t, ix.Fields),
				"update", inflect.Singularize(t.Name), "by", joinField(ix.Fields, "And"), "stmt",
			)
			writeConst(nil, w,
				d.Delete(t, ix.Fields),
				"delete", inflect.Singularize(t.Name), "by", joinField(ix.Fields, "And"), "stmt",
			)
		}
	}

	for _, fk := range t.Foreigns{
		writeConst(sqlFileContent, w,
			d.Foreign(t, fk),
			"create", inflect.Singularize(fk.Name), "stmt",
		)

		writeConst(nil, w,
			d.Select(t, fk.FromFields),
			"select", inflect.Singularize(t.Name), "of", inflect.Singularize(fk.ToTable), "by", joinColumnNames(fk.FromColumns, "And"), "stmt",
		)

		if fk.Many{
			writeConst(nil, w,
				d.SelectRange(t, fk.FromFields),
				"select", inflect.Singularize(t.Name), "of", inflect.Singularize(fk.ToTable), "range", "by", joinField(fk.FromFields, "And"), "stmt",
			)
		}
	}

	if outputSql{
		sqlFile.Write(sqlFileContent.Bytes())
	}
}

// WritePackage writes the Go package header to
// writer w with the given package name.
func writePackage(w io.Writer, name string) {
	fmt.Fprintf(w, sPackage, name)
}

// writeConst is a helper function that writes the
// body string to a Go const variable.
func writeConst(content *bytes.Buffer, w io.Writer, body string, label ...string) string{
	// create a snake case variable name from
	// the specified labels. Then convert the
	// variable name to a quoted, camel case string.
	name := getLabelName(label...)

	// quote the body using multi-line quotes
	body = fmt.Sprintf(sQuote, body)

	if content!=nil{
		content.WriteString(body[1:len(body)-1])
	}
	fmt.Fprintf(w, sConst, name, body)
	log.Printf("const name:%s",name)
	return name
}

func getLabelName(label ...string) string{
	name := strings.Join(label, "_")
	name = inflect.Typeify(name)
	if strings.HasSuffix(name, `Stmt`){
		name = inflect.CamelizeDownFirst(name)
	}
	return name
}

func joinField(fields[]*schema.Field, sep string)string{
	var buf bytes.Buffer
	for i,field := range fields{
		if i==0{
			buf.WriteString(inflect.Camelize(field.Name[2:]))
		}else{
			buf.WriteString(sep)
			buf.WriteString(inflect.Camelize(field.Name[2:]))
		}
	}
	return buf.String()
}

func joinColumnNames(fields[]string, sep string)string{
	var buf bytes.Buffer
	for i,field := range fields{
		if i==0{
			buf.WriteString(inflect.Camelize(field[2:]))
		}else{
			buf.WriteString(sep)
			buf.WriteString(inflect.Camelize(field[2:]))
		}
	}
	return buf.String()
}

func joinObjectField(fields[]*schema.Field, sep string)string{
	var buf bytes.Buffer
	for i,field := range fields{
		if i==0{
			buf.WriteString(`v.`+field.Node.Name)
		}else{
			buf.WriteString(sep)
			buf.WriteString(`v.`+field.Node.Name)
		}
	}
	return buf.String()
}

func joinObjectFieldInDetails(fields[]*schema.Field, sep string, withType bool)string{
	var buf bytes.Buffer
	for i,field := range fields{
		if i==0{
			buf.WriteString(inflect.CamelizeDownFirst(field.Node.Name))
			if withType{
				buf.WriteString(" "+field.Node.Type)
			}
		}else{
			buf.WriteString(sep)
			buf.WriteString(inflect.CamelizeDownFirst(field.Node.Name))
			if withType{
				buf.WriteString(" "+field.Node.Type)
			}
		}
	}
	return buf.String()
}