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

	if *genFuncs {
		if *needImport{
			writePackage(&buf, *pkgName)
			writeImports(&buf, tree, "database/sql", *srcPkgName)
		}
		writeRowFunc(srcPkgNameInShort, &buf, tree)
		writeRowsFunc(srcPkgNameInShort, &buf, tree)
		writeSliceFunc(srcPkgNameInShort, &buf, tree)

		if *extraFuncs {
			writeSelectRow(srcPkgNameInShort, &buf, tree)
			writeSelectRows(srcPkgNameInShort, &buf, tree)
			writeInsertFunc(srcPkgNameInShort, &buf, tree)
			writeUpdateFunc(srcPkgNameInShort, &buf, tree)
		}
	} else {
		writePackage(&buf, *pkgName)
	}

	// write the sql functions
	if *genSchema {
		writeSchema(&buf, dialect, table, *outputSql)
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
