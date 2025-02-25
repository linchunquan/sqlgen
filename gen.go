package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/linchunquan/sqlgen/parse"
	"github.com/linchunquan/sqlgen/schema"
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

	log.Printf("Finish write the sql functions for table %s\n", table.Name)

	if *genFuncs {

		writeRowFunc(srcPkgNameInShort, &buf, tree)
		log.Printf("Finish writeRowFunc for table %s\n", table.Name)
		writeRowsFunc(srcPkgNameInShort, &buf, tree)
		log.Printf("Finish writeRowsFunc for table %s\n", table.Name)
		writeSliceFunc(srcPkgNameInShort, &buf, tree)
		log.Printf("Finish writeSliceFunc for table %s\n", table.Name)

		if *extraFuncs {
			writeGenericSelectRow(srcPkgNameInShort, &buf, tree)
			log.Printf("Finish writeGenericSelectRow for table %s\n", table.Name)
			writeGenericSelectRows(srcPkgNameInShort, &buf, tree)
			log.Printf("Finish writeGenericSelectRows for table %s\n", table.Name)
			//writeGenericInsertFunc(srcPkgNameInShort, &buf, tree)
			//writeGenericUpdateFunc(srcPkgNameInShort, &buf, tree)
			if !*view {
				writeInsertFunc(srcPkgNameInShort, &buf, tree, table)
				log.Printf("Finish writeInsertFunc for table %s\n", table.Name)
				writeDeleteFunc(srcPkgNameInShort, &buf, tree, table)
				log.Printf("Finish writeDeleteFunc for table %s\n", table.Name)
				writeUpdateFunc(srcPkgNameInShort, &buf, tree, table)
				log.Printf("Finish writeUpdateFunc for table %s\n", table.Name)
			}
			writeGetByFunc(srcPkgNameInShort, &buf, tree, table)
			log.Printf("Finish writeGetByFunc for table %s\n", table.Name)
			writeFindAllFunc(srcPkgNameInShort, &buf, tree, table)
			log.Printf("Finish writeFindAllFunc for table %s\n", table.Name)
			writeFindAllInRangeFunc(srcPkgNameInShort, &buf, tree, table)
			log.Printf("Finish writeFindAllInRangeFunc for table %s\n", table.Name)
			writeFindByIndexFunc(srcPkgNameInShort, &buf, tree, table)
			log.Printf("Finish writeFindByIndexFunc for table %s\n", table.Name)
			writeFindByForeignKeyFunc(srcPkgNameInShort, &buf, tree, table)
			log.Printf("Finish writeFindByForeignKeyFunc for table %s\n", table.Name)
			writeCountAllFunc(&buf,tree,table)
			log.Printf("Finish writeCountAllFunc for table %s\n", table.Name)
			writeCountByIndexFunc(&buf,tree,table)
			log.Printf("Finish writeCountByIndexFunc for table %s\n", table.Name)
		}
	} else {
		writePackage(&buf, *pkgName)
		log.Printf("Finish writePackage for table %s\n", table.Name)
	}

	log.Printf("Generate content for table %s\n", table.Name)
	log.Println("==================================================================")
	log.Printf("%s\n", buf.Bytes())
	log.Println("==================================================================")

	// formats the generated file using gofmt
	pretty, err := format(&buf)
	log.Printf("Finish format for table %s\n, err:%v\n", table.Name, err)
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
