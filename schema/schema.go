package schema

import(
	"github.com/linchunquan/sqlgen/parse"
)
// List of basic types
const (
	INTEGER int = iota
	VARCHAR
	BOOLEAN
	REAL
	BLOB
)

// List of vendor-specific keywords
const (
	AUTO_INCREMENT = iota
	PRIMARY_KEY
)

type Table struct {
	Name string

	Fields  []*Field
	Index   []*Index
	Primary []*Field
	Foreigns[]*Foreign
}

type Field struct {
	Node    *parse.Node
	Name    string
	Type    int
	Primary bool
	Auto    bool
	Size    int
}

type Index struct {
	Name   string
	Unique bool

	Fields []*Field
}

type Foreign struct{
	Name      string
	FromColumns []string
	ToTable     string
	ToColumns   []string
}
