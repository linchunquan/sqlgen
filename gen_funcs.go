package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"bitbucket.org/pkg/inflect"
	"github.com/acsellers/inflections"
	"github.com/linchunquan/sqlgen/parse"
	"github.com/linchunquan/sqlgen/schema"
)

func writeImports(w io.Writer, tree *parse.Node, pkgs ...string) {
	var pmap = map[string]struct{}{}

	// add default packages
	for _, pkg := range pkgs {
		pmap[pkg] = struct{}{}
	}

	// check each edge node to see if it is
	// encoded, which might require us to import
	// other packages
	for _, node := range tree.Edges() {
		if node.Tags == nil || len(node.Tags.Encode) == 0 {
			continue
		}
		switch node.Tags.Encode {
		case "json":
			pmap["encoding/json"] = struct{}{}
			// case "gzip":
			// 	pmap["compress/gzip"] = struct{}{}
			// case "snappy":
			// 	pmap["github.com/golang/snappy"] = struct{}{}
		}
	}

	if len(pmap) == 0 {
		return
	}

	// write the import block, including each
	// encoder package that was specified.
	fmt.Fprintln(w, "\nimport (")
	for pkg, _ := range pmap {
		fmt.Fprintf(w, "\t%q\n", pkg)
	}
	fmt.Fprintln(w, ")")
}

func writeSliceFunc(srcPkgNameInShort string, w io.Writer, tree *parse.Node) {

	var buf1, buf2, buf3 bytes.Buffer

	var i, depth int
	var parent = tree

	for _, node := range tree.Edges() {
		if node.Tags.Skip {
			continue
		}

		// temporary variable declaration
		switch node.Kind {
		case parse.Map, parse.Slice:
			fmt.Fprintf(&buf1, "var v%d %s\n", i, "[]byte")
		default:
			fmt.Fprintf(&buf1, "var v%d %s\n", i, node.Type)
		}

		// variable scanning
		fmt.Fprintf(&buf3, "v%d,\n", i)

		// variable setting
		path := node.Path()[1:]

		// if the parent is a ptr struct we
		// need to create a new
		if parent != node.Parent && node.Parent.Kind == parse.Ptr {
			// if node.Parent != nil && node.Parent.Parent != parent {
			// 	fmt.Fprintln(&buf2, "}\n")
			// 	depth--
			// }

			// seriously ... this works?
			if node.Parent != nil && node.Parent.Parent != parent {
				for _, p := range path {
					if p == parent || depth == 0 {
						break
					}
					fmt.Fprintln(&buf2, "}\n")
					depth--
				}
			}
			depth++
			fmt.Fprintf(&buf2, "if v.%s != nil {\n", join(path[:len(path)-1], "."))
		}

		switch node.Kind {
		case parse.Map, parse.Slice, parse.Struct, parse.Ptr:
			fmt.Fprintf(&buf2, "v%d, _ = json.Marshal(&v.%s)\n", i, join(path, "."))
		default:
			fmt.Fprintf(&buf2, "v%d=v.%s\n", i, join(path, "."))
		}

		parent = node.Parent
		i++
	}

	for depth != 0 {
		depth--
		fmt.Fprintln(&buf2, "}\n")
	}

	fmt.Fprintf(w,
		sSliceRow,
		tree.Type,
		srcPkgNameInShort+"."+tree.Type,
		buf1.String(),
		buf2.String(),
		buf3.String(),
	)
}

func getAssignmentCode(buf *bytes.Buffer, node *parse.Node, i int, attr string) {
	tmp := `
    if v%d.Valid{
        v.%s=v%d.%s
    }else{
        v.%s=%s
    }
`
	if node.Type == "int" {
		tmp = `
    if v%d.Valid{
        v.%s=int(v%d.%s64)
    }else{
        v.%s=%s
    }
`
	}

	value := strings.Title(node.Type)
	defautlVal := `""`
	if strings.Contains(node.Type, "bool") {
		defautlVal = "false"
	} else if strings.Contains(node.Type, "float") {
		defautlVal = "0"
	} else if strings.Contains(node.Type, "int") {
		defautlVal = "0"
	} else if strings.Contains(node.Type, "[]byte") {
		defautlVal = "nil"
		value = "Bytes"
	}

	fmt.Fprintf(buf, tmp, i, attr, i, value, attr, defautlVal)
}

func getSqlNullType(node *parse.Node) string{
	if node.Type == "int" {
		return "sql.NullInt64"
	} else if node.Type == "[]byte" {
		return "db.NullBytes"
	}
	return "sql.Null"+strings.Title(node.Type)
}
func writeRowFunc(srcPkgNameInShort string, w io.Writer, tree *parse.Node) {

	var buf1, buf2, buf3 bytes.Buffer

	var i int
	var parent = tree
	for _, node := range tree.Edges() {
		if node.Tags.Skip {
			continue
		}

		// temporary variable declaration
		switch node.Kind {
		case parse.Map, parse.Slice:
			fmt.Fprintf(&buf1, "var v%d %s\n", i, "[]byte")
		default:
			fmt.Fprintf(&buf1, "var v%d %s\n", i, getSqlNullType(node))
		}

		// variable scanning
		fmt.Fprintf(&buf2, "&v%d,\n", i)

		// variable setting
		path := node.Path()[1:]

		// if the parent is a ptr struct we
		// need to create a new
		if parent != node.Parent && node.Parent.Kind == parse.Ptr {
			fmt.Fprintf(&buf3, "v.%s=&%s{}\n", join(path[:len(path)-1], "."), node.Parent.Type)
		}

		switch node.Kind {
		case parse.Map, parse.Slice, parse.Struct, parse.Ptr:
			fmt.Fprintf(&buf3, "json.Unmarshal(v%d, &v.%s)\n", i, join(path, "."))
		default:
			//fmt.Fprintf(&buf3, "v.%s=v%d\n", join(path, "."), i)
			getAssignmentCode(&buf3, node, i, join(path, "."))
		}

		parent = node.Parent
		i++
	}
	//fmt.Printf("tree.Type:%v",tree.Type)
	fmt.Fprintf(w,
		sScanRow,
		//fmt.Sprintf("%s.%s", tree.Type),
		tree.Type,
		srcPkgNameInShort+"."+tree.Type,
		buf1.String(),
		buf2.String(),
		srcPkgNameInShort+"."+tree.Type,
		buf3.String(),
	)
}

func writeRowsFunc(srcPkgNameInShort string, w io.Writer, tree *parse.Node) {
	var buf1, buf2, buf3 bytes.Buffer

	var i int
	var parent = tree
	for _, node := range tree.Edges() {
		if node.Tags.Skip {
			continue
		}

		// temporary variable declaration
		switch node.Kind {
		case parse.Map, parse.Slice:
			fmt.Fprintf(&buf1, "var v%d %s\n", i, "[]byte")
		default:
			fmt.Fprintf(&buf1, "var v%d %s\n", i, getSqlNullType(node))
		}

		// variable scanning
		fmt.Fprintf(&buf2, "&v%d,\n", i)

		// variable setting
		path := node.Path()[1:]

		// if the parent is a ptr struct we
		// need to create a new
		if parent != node.Parent && node.Parent.Kind == parse.Ptr {
			fmt.Fprintf(&buf3, "v.%s=&%s{}\n", join(path[:len(path)-1], "."), node.Parent.Type)
		}

		switch node.Kind {
		case parse.Map, parse.Slice, parse.Struct, parse.Ptr:
			fmt.Fprintf(&buf3, "json.Unmarshal(v%d, &v.%s)\n", i, join(path, "."))
		default:
			//fmt.Fprintf(&buf3, "v.%s=v%d\n", join(path, "."), i)
			getAssignmentCode(&buf3, node, i, join(path, "."))
		}

		parent = node.Parent
		i++
	}

	fmt.Fprintf(w,
		sScanRows,
		inflections.Pluralize(tree.Type),
		srcPkgNameInShort+"."+tree.Type,
		srcPkgNameInShort+"."+tree.Type,
		buf1.String(),
		buf2.String(),
		srcPkgNameInShort+"."+tree.Type,
		buf3.String(),
	)
}

func writeGenericSelectRow(srcPkgNameInShort string, w io.Writer, tree *parse.Node) {
	fmt.Fprintf(w, sGenericSelectRow, tree.Type, srcPkgNameInShort+"."+tree.Type, tree.Type)
}

func writeGenericSelectRows(srcPkgNameInShort string,w io.Writer, tree *parse.Node) {
	plural := inflections.Pluralize(tree.Type)
	fmt.Fprintf(w, sGenericSelectRows, plural, srcPkgNameInShort+"."+tree.Type, plural)
}

func writeGenericInsertFunc(srcPkgNameInShort string, w io.Writer, tree *parse.Node) {
	// TODO this assumes I'm using the ID field.
	// we should not make that assumption
	fmt.Fprintf(w, sGenericInsert, tree.Type, srcPkgNameInShort+"."+tree.Type, tree.Type)
}

func writeGenericUpdateFunc(srcPkgNameInShort string, w io.Writer, tree *parse.Node) {
	fmt.Fprintf(w, sGenericUpdate, tree.Type, srcPkgNameInShort+"."+tree.Type, tree.Type)
}

func writeInsertFunc(srcPkgNameInShort string, w io.Writer,  tree *parse.Node, t *schema.Table){
	fmt.Fprintf(w, sInsert, tree.Type, srcPkgNameInShort+"."+tree.Type, getLabelName("insert", inflect.Singularize(t.Name), "stmt"), tree.Type)
}

func writeDeleteFunc(srcPkgNameInShort string, w io.Writer,  tree *parse.Node, t *schema.Table){
	if len(t.Primary) !=0 {
		fmt.Fprintf(w, sDelete,
			tree.Type,
			getLabelName("by", joinField(t.Primary, "And")),
			joinObjectFieldInDetails(t.Primary, ",", true),
			joinObjectFieldInDetails(t.Primary, ",", false),
			getLabelName("delete", inflect.Singularize(t.Name), "by", joinField(t.Primary, "And"), "stmt"))
	}
	if len(t.Index) !=0 {
		for _, ix := range t.Index {
			//if ix.Unique {
				fmt.Fprintf(w, sDelete,
					tree.Type,
					getLabelName("by", joinField(ix.Fields, "And")),
					joinObjectFieldInDetails(ix.Fields, ",", true),
					joinObjectFieldInDetails(ix.Fields, ",", false),
					getLabelName("delete", inflect.Singularize(t.Name), "by", joinField(ix.Fields, "And"), "stmt"))
			//}
		}
	}
}

func writeUpdateFunc(srcPkgNameInShort string, w io.Writer,  tree *parse.Node, t *schema.Table){
	if len(t.Primary) !=0 {
		fmt.Fprintf(w, sUpdate,
			tree.Type,
			getLabelName("by", joinField(t.Primary, "And")),
			srcPkgNameInShort+"."+tree.Type,
			tree.Type,
			joinObjectField(t.Primary, ","),
			getLabelName("update", inflect.Singularize(t.Name), "by", joinField(t.Primary, "And"), "stmt"))
	}
	if len(t.Index) !=0 {
		for _, ix := range t.Index {
			if ix.Unique {
				fmt.Fprintf(w, sUpdate,
					tree.Type,
					getLabelName("by", joinField(ix.Fields, "And")),
					srcPkgNameInShort+"."+tree.Type,
					tree.Type,
					joinObjectField(ix.Fields, ","),
					getLabelName("update", inflect.Singularize(t.Name), "by", joinField(ix.Fields, "And"), "stmt"))
			}
		}
	}
}

func writeGetByFunc(srcPkgNameInShort string, w io.Writer,  tree *parse.Node, t *schema.Table){
	if len(t.Primary) !=0 {
		fmt.Fprintf(w, sGetBy,
			tree.Type,
			getLabelName("by", joinField(t.Primary, "And")),
			joinObjectFieldInDetails(t.Primary, ",", true),
			srcPkgNameInShort+"."+tree.Type,
			joinObjectFieldInDetails(t.Primary, ",", false),
			tree.Type,
			getLabelName("select", inflect.Singularize(t.Name), "by", joinField(t.Primary, "And"), "stmt"))
	}
	if len(t.Index) !=0 {
		for _, ix := range t.Index {
			if ix.Unique {
				fmt.Fprintf(w, sGetBy,
					tree.Type,
					getLabelName("by", joinField(ix.Fields, "And")),
					joinObjectFieldInDetails(ix.Fields, ",", true),
					srcPkgNameInShort+"."+tree.Type,
					joinObjectFieldInDetails(ix.Fields, ",", false),
					tree.Type,
					getLabelName("select", inflect.Singularize(t.Name), "by", joinField(ix.Fields, "And"), "stmt"))
			}
		}
	}
}

func writeFindByIndexFunc(srcPkgNameInShort string, w io.Writer,  tree *parse.Node, t *schema.Table){
	if len(t.Index) !=0 {
		for _, ix := range t.Index {
			if !ix.Unique {
				fmt.Fprintf(w, sFindByIndex,
					tree.Type,
					getLabelName("by", joinField(ix.Fields, "And")),
					joinObjectFieldInDetails(ix.Fields, ",", true),
					srcPkgNameInShort+"."+tree.Type,
					joinObjectFieldInDetails(ix.Fields, ",", false),
					tree.Type,
					getLabelName("select", inflect.Singularize(t.Name), "by", joinField(ix.Fields, "And"), "stmt"))

				fmt.Fprintf(w, sFindByIndexInRange,
					tree.Type,
					getLabelName("by", joinField(ix.Fields, "And")),
					joinObjectFieldInDetails(ix.Fields, ",", true),
					srcPkgNameInShort+"."+tree.Type,
					joinObjectFieldInDetails(ix.Fields, ",", false),
					tree.Type,
					getLabelName("select", inflect.Singularize(t.Name), "range", "by", joinField(ix.Fields, "And"), "stmt"))
			}
		}
	}
}

func writeFindByForeignKeyFunc(srcPkgNameInShort string, w io.Writer,  tree *parse.Node, t *schema.Table){
	if len(t.Foreigns) !=0 {
		for _, fk := range t.Foreigns {
			if fk.Many {
				fmt.Fprintf(w, sFindByForeignKey,
					tree.Type,
					inflect.Camelize(fk.ToTable[:len(fk.ToTable)-1]),
					getLabelName("by", joinField(fk.FromFields, "And")),
					joinObjectFieldInDetails(fk.FromFields, ",", true),
					srcPkgNameInShort+"."+tree.Type,
					joinObjectFieldInDetails(fk.FromFields, ",", false),
					tree.Type,
					getLabelName("select", inflect.Singularize(t.Name), "of", inflect.Singularize(fk.ToTable), "by", joinColumnNames(fk.FromColumns, "And"), "stmt"))

				fmt.Fprintf(w, sFindByForeignKeyInRange,
					tree.Type,
					inflect.Camelize(fk.ToTable[:len(fk.ToTable)-1]),
					getLabelName("by", joinField(fk.FromFields, "And")),
					joinObjectFieldInDetails(fk.FromFields, ",", true),
					srcPkgNameInShort+"."+tree.Type,
					joinObjectFieldInDetails(fk.FromFields, ",", false),
					tree.Type,
					getLabelName("select", inflect.Singularize(t.Name), "of", inflect.Singularize(fk.ToTable), "range", "by", joinField(fk.FromFields, "And"), "stmt"))
			}else{
				fmt.Fprintf(w, sGetByForeignKey,
					tree.Type,
					inflect.Camelize(fk.ToTable[:len(fk.ToTable)-1]),
					getLabelName("by", joinField(fk.FromFields, "And")),
					joinObjectFieldInDetails(fk.FromFields, ",", true),
					srcPkgNameInShort+"."+tree.Type,
					joinObjectFieldInDetails(fk.FromFields, ",", false),
					tree.Type,
					getLabelName("select", inflect.Singularize(t.Name), "of", inflect.Singularize(fk.ToTable), "by", joinColumnNames(fk.FromColumns, "And"), "stmt"))
			}
		}
	}
}

func writeFindAllFunc(srcPkgNameInShort string, w io.Writer,  tree *parse.Node, t *schema.Table){
	fmt.Fprintf(w, sFindAll, tree.Type, srcPkgNameInShort+"."+tree.Type, tree.Type, getLabelName("select", inflect.Singularize(t.Name), "stmt"))
}

func writeFindAllInRangeFunc(srcPkgNameInShort string, w io.Writer,  tree *parse.Node, t *schema.Table){
	fmt.Fprintf(w, sFindAllInRange, tree.Type, srcPkgNameInShort+"."+tree.Type, tree.Type, getLabelName("select", inflect.Singularize(t.Name), "range", "stmt"))
}

func writeCountAllFunc(w io.Writer,  tree *parse.Node, t *schema.Table){
	fmt.Fprintf(w, sCount, tree.Type,  getLabelName("select", inflect.Singularize(t.Name), "count", "stmt"))
}

func writeCountByIndexFunc(w io.Writer,  tree *parse.Node, t *schema.Table){
	if len(t.Index) !=0 {
		for _, ix := range t.Index {
			fmt.Fprintf(w, sCountByIndex,
				tree.Type,
				getLabelName("by", joinField(ix.Fields, "And")),
				joinObjectFieldInDetails(ix.Fields, ",", true),
				joinObjectFieldInDetails(ix.Fields, ",", false),
				getLabelName("select", inflect.Singularize(t.Name), "count", "by", joinField(ix.Fields, "And"), "stmt"))
		}
	}
}

// join is a helper function that joins nodes
// together by name using the seperator.
func join(nodes []*parse.Node, sep string) string {
	var parts []string
	for _, node := range nodes {
		parts = append(parts, node.Name)
	}
	return strings.Join(parts, sep)
}
