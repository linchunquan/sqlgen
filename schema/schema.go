package schema

import (
	"github.com/linchunquan/sqlgen/parse"
)

// List of basic types
const (
	INTEGER int = iota
	LONG
	VARCHAR
	BOOLEAN
	REAL
	BLOB
	FLOAT
	DOUBLE
	MEDIUMTEXT
	LONGTEXT
)

// List of vendor-specific keywords
const (
	AUTO_INCREMENT = iota
	PRIMARY_KEY
)

type Table struct {
	Name     string
	Fields   []*Field
	Index    []*Index
	Primary  []*Field
	Foreigns []*Foreign
}

type Field struct {
	Node    *parse.Node
	Name    string
	Type    int
	Primary bool
	Auto    bool
	Size    int
	Operator string
	ValueAsFirstArg bool
}

func(f*Field)Clone()*Field{
	return &Field{Node:f.Node, Name:f.Name, Type:f.Type, Primary:f.Primary, Auto:f.Auto, Size:f.Size, Operator:f.Operator, ValueAsFirstArg:f.ValueAsFirstArg}
}

type Index struct {
	Name    string
	Unique  bool
	Fields  []*Field
}

type Foreign struct{
	Name        string
	FromColumns []string
	FromFields  []*Field
	ToTable     string
	ToColumns   []string
	Many        bool
}