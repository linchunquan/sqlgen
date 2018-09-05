package schema

import (
	"strings"

	"github.com/acsellers/inflections"
	"github.com/linchunquan/sqlgen/parse"
	"log"
)

func Load(tree *parse.Node) *Table {
	table := new(Table)

	// local map of indexes, used for quick
	// lookups and de-duping.
	indexs := map[string]*Index{}
	foreigns := map[string]*Foreign{}

	// pluralizes the table name and then
	// formats in snake case.
	table.Name = inflections.Underscore(tree.Type)
	table.Name = inflections.Pluralize(table.Name)

	// each edge node in the tree is a column
	// in the table. Convert each edge node to
	// a Field structure.
	for _, node := range tree.Edges() {

		field := new(Field)

		// Lookup the SQL column type
		// TODO: move this to a function
		t, ok := parse.Types[node.Type]
		if ok {
			tt, ok := types[t]
			if !ok {
				tt = BLOB
			}
			field.Type = tt
		} else {
			field.Type = BLOB
		}

		// get the full path name
		path := node.Path()
		var parts []string
		for _, part := range path {
			if part.Tags != nil && part.Tags.Name != "" {
				parts = append(parts, part.Tags.Name)
				continue
			}

			parts = append(parts, part.Name)
		}

		//fix_here, simplify the field name
		parts[0]="f"

		field.Node = node
		field.Name = strings.Join(parts, "_")
		field.Name = inflections.Underscore(field.Name)

		// substitute tag variables
		if node.Tags != nil {

			if node.Tags.Skip {
				continue
			}

			// default ID and int64 to primary key
			// with auto-increment
			if node.Name == "ID" && node.Kind == parse.Int64 {
				node.Tags.Primary = true
				node.Tags.Auto = true
			}

			field.Auto = node.Tags.Auto
			field.Primary = node.Tags.Primary
			field.Size = node.Tags.Size

			if node.Tags.Primary {
				table.Primary = append(table.Primary, field)
			}

			if node.Tags.Index != "" {
				index, ok := indexs[node.Tags.Index]
				if !ok {
					index = new(Index)
					index.Name = node.Tags.Index
					indexs[index.Name] = index
					table.Index = append(table.Index, index)
				}
				index.Fields = append(index.Fields, field)
			}

			if node.Tags.Unique != "" {
				index, ok := indexs[node.Tags.Unique]
				if !ok {
					index = new(Index)
					index.Name = node.Tags.Unique
					index.Unique = true
					indexs[index.Name] = index
					table.Index = append(table.Index, index)
				}
				index.Unique = true
				index.Fields = append(index.Fields, field)
			}

			if node.Tags.Type != "" {
				t, ok := sqlTypes[node.Tags.Type]
				if ok {
					field.Type = t
				}
			}

			if node.Tags.Foreign != "" {
				strs := strings.Split(node.Tags.Foreign, "@")
				if len(strs)==2{

					tableName := strings.TrimSpace(strs[1])

					if len(tableName)>0{

						var fkName string
						if len(node.Tags.ForeignGroup)==0{
							fkName = "fk_"+table.Name+"_to_"+tableName
						}else{
							fkName = node.Tags.ForeignGroup
						}

						foreign,ok := foreigns[fkName]
						if !ok {
							foreign = new(Foreign)
							foreign.Name = fkName
							foreign.Many = node.Tags.Many
							foreign.ToTable = tableName
							foreigns[fkName] = foreign
							table.Foreigns = append(table.Foreigns, foreign)
							log.Printf("add foreign key:%+v", foreign)
						}

						foreign.FromColumns = append(foreign.FromColumns, field.Name)
						foreign.FromFields = append(foreign.FromFields, field)
						foreign.ToColumns = append(foreign.ToColumns, "f_"+strings.TrimSpace(inflections.Underscore(strs[0])))
					}
				}
			}
		}

		table.Fields = append(table.Fields, field)
	}

	return table
}

// convert Go types to SQL types.
var types = map[uint8]int{
	parse.Bool:       BOOLEAN,
	parse.Int:        INTEGER,
	parse.Int8:       INTEGER,
	parse.Int16:      INTEGER,
	parse.Int32:      INTEGER,
	parse.Int64:      LONG,
	parse.Uint:       INTEGER,
	parse.Uint8:      INTEGER,
	parse.Uint16:     INTEGER,
	parse.Uint32:     INTEGER,
	parse.Uint64:     INTEGER,
	parse.Float32:    INTEGER,
	parse.Float64:    INTEGER,
	parse.Complex64:  INTEGER,
	parse.Complex128: INTEGER,
	parse.Interface:  BLOB,
	parse.Bytes:      BLOB,
	parse.String:     VARCHAR,
	parse.Map:        BLOB,
	parse.Slice:      BLOB,
}

var sqlTypes = map[string]int{
	"text":     VARCHAR,
	"varchar":  VARCHAR,
	"varchar2": VARCHAR,
	"number":   INTEGER,
	"integer":  INTEGER,
	"int":      INTEGER,
	"long":     LONG,
	"blob":     BLOB,
	"bytea":    BLOB,
}
