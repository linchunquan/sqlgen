package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"github.com/linchunquan/sqlgen/parse"
	"github.com/linchunquan/sqlgen/schema"
	"log"
)

var (
	input      = flag.String("file", "", "input file name; required")
	output     = flag.String("o", "", "output file name; required")
	outputSql  = flag.String("osf", "", "output sql file path;")
	pkgName    = flag.String("pkg", "main", "output package name; required")
	srcPkgName = flag.String("srcPkg", "main", "input package name; required")
	typeName   = flag.String("type", "", "type to generate; required")
	database   = flag.String("db", "sqlite", "sql dialect; required")
	genSchema  = flag.Bool("schema", true, "generate sql schema and queries")
	genFuncs   = flag.Bool("funcs", true, "generate sql helper functions")
	extraFuncs = flag.Bool("extras", true, "generate extra sql helper functions")
	needImport = flag.Bool( "needImport", true, "need to generate import statement")
	view       = flag.Bool("view", false, "is view, not table")
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	flag.Parse()

	// parses the syntax tree into something a bit
	// easier to work with.
	tree, err := parse.Parse(*input, *typeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// if the code is generated in a different folder
	// that the struct we need to import the struct
	if tree.Pkg != *pkgName && *pkgName != "main" {
		// TODO
	}

	// load the Tree into a schema Object
	table := schema.Load(tree)
	dialect := schema.New(schema.Dialects[*database])
	strs:=strings.Split(*srcPkgName, "/")
	srcPkgNameInShort:=strs[len(strs)-1]


	var buf bytes.Buffer

	if *needImport{
		writePackage(&buf, *pkgName)
		writeImports(&buf, tree, "database/sql", "github.com/linchunquan/sqlgen/db", *srcPkgName)
	}

	// write the sql functions
	if *genSchema {
		writeSchema(&buf, dialect, table, *outputSql, *view)
	}

	if *genFuncs {

		writeRowFunc(srcPkgNameInShort, &buf, tree)
		writeRowsFunc(srcPkgNameInShort, &buf, tree)
		writeSliceFunc(srcPkgNameInShort, &buf, tree)

		if *extraFuncs {
			writeGenericSelectRow(srcPkgNameInShort, &buf, tree)
			writeGenericSelectRows(srcPkgNameInShort, &buf, tree)
			//writeGenericInsertFunc(srcPkgNameInShort, &buf, tree)
			//writeGenericUpdateFunc(srcPkgNameInShort, &buf, tree)
			if !*view {
				writeInsertFunc(srcPkgNameInShort, &buf, tree, table)
				writeDeleteFunc(srcPkgNameInShort, &buf, tree, table)
				writeUpdateFunc(srcPkgNameInShort, &buf, tree, table)
			}
			writeGetByFunc(srcPkgNameInShort, &buf, tree, table)
			writeFindAllFunc(srcPkgNameInShort, &buf, tree, table)
			writeFindAllInRangeFunc(srcPkgNameInShort, &buf, tree, table)
			writeFindByIndexFunc(srcPkgNameInShort, &buf, tree, table)
			writeFindByForeignKeyFunc(srcPkgNameInShort, &buf, tree, table)
			writeCountAllFunc(&buf,tree,table)
			writeCountByIndexFunc(&buf,tree,table)
		}
	} else {
		writePackage(&buf, *pkgName)
	}

	// formats the generated file using gofmt
	pretty, err := format(&buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	// create output source for file. defaults to
	// stdout but may be file.
	var out io.WriteCloser = os.Stdout
	if *output != "" {
		out, err = os.Create(*output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}
		defer out.Close()
	}

	io.Copy(out, pretty)
}
