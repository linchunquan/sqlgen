package schema

import (
	"log"
	"strings"

	"github.com/acsellers/inflections"
	"github.com/linchunquan/sqlgen/parse"
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

			if node.Tags.TableName != "" {
				log.Printf("set table name as %s \n", node.Tags.TableName)
				table.Name = node.Tags.TableName
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
				indexSet, notEmpty := splitIndexString(node.Tags.Index)
				if notEmpty{
					for i, _ := range indexSet{
						idxInfo:=indexSet[i]
						indexName:=idxInfo.name
						index, ok := indexs[indexName]
						if !ok {
							index = new(Index)
							index.Name = indexName
							indexs[index.Name] = index
							table.Index = append(table.Index, index)
						}
						idxField := field.Clone()
						idxField.Operator = idxInfo.operator
						idxField.ValueAsFirstArg = idxInfo.valueAsFirstArg
						index.Fields = append(index.Fields, idxField)
					}
				}
			}

			if node.Tags.Unique != "" {
				indexSet, notEmpty := splitIndexString(node.Tags.Unique)
				if notEmpty{
					for i, _ := range indexSet{
						idxInfo:=indexSet[i]
						indexName:=idxInfo.name
						index, ok := indexs[indexName]
						if !ok {
							index = new(Index)
							index.Name = indexName
							index.Unique = true
							indexs[index.Name] = index
							table.Index = append(table.Index, index)
						}
						index.Unique = true
						idxField := field.Clone()
						idxField.Operator = idxInfo.operator
						idxField.ValueAsFirstArg = idxInfo.valueAsFirstArg
						index.Fields = append(index.Fields, idxField)
					}
				}
			}

			if node.Tags.Type != "" {
				t, ok := sqlTypes[node.Tags.Type]
				if ok {
					field.Type = t
				}
			}

			if node.Tags.Foreign != "" {
				foreignConfigs := strings.Split(node.Tags.Foreign, ";")
				n:=len(foreignConfigs)
				for i:=0;i<n;i++{
					strs := strings.Split(strings.TrimSpace(foreignConfigs[i]), "@")
					if len(strs)>=2{
						tableName := strings.TrimSpace(strs[1])
						if len(tableName)>0{
							var fkName string
							if len(strs)==2{
								if len(node.Tags.ForeignGroup)==0{
									fkName = "fk_"+table.Name+"_to_"+tableName
								}else{
									fkName = node.Tags.ForeignGroup
								}
							}else{
								fkName = strs[2]
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
		}

		table.Fields = append(table.Fields, field)
	}

	return table
}

type indexInfo struct{
	name string
	operator string
	valueAsFirstArg bool
}

func splitIndexString(indexString string)(indexSet[]*indexInfo, notEmpty bool){
	strs := strings.Split(strings.TrimSpace(indexString), ";")
	n:=len(strs)
	for i:=0;i<n;i++{
		str := strings.TrimSpace(strings.Replace(strs[i]," ","", -1))
		if len(str)>0{
			strs2 := strings.Split(str, "@")
			name := strs2[0]
			if len(name)>0{
				idx := &indexInfo{name:name}
				if len(strs2)>1{
					idx.operator=strings.TrimSpace(strs2[1])
				}
				if len(strs2)>2{
					if strings.EqualFold("TRUE", strings.ToUpper(strs2[2])){
						idx.valueAsFirstArg = true
					}
				}
				indexSet = append(indexSet, idx)
			}
		}
	}
	notEmpty = len(indexSet)>0
	return indexSet, notEmpty
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
	parse.Float32:    FLOAT,
	parse.Float64:    DOUBLE,
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
	"double":   DOUBLE,
	"float":    FLOAT,
	"MEDIUMTEXT": MEDIUMTEXT,
	"LONGTEXT": LONGTEXT,
}
