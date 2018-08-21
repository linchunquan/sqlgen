package schema

const (
	SQLITE int = iota
	POSTGRES
	MYSQL
)

var Dialects = map[string]int{
	"sqlite":   SQLITE,
	"postgres": POSTGRES,
	"mysql":    MYSQL,
}

type Dialect interface {
	Table(*Table) string
	Index(*Table, *Index) string
	Foreign(*Table, *Foreign) string
	Column(*Field) string
	Insert(*Table) string
	Update(*Table, []*Field) string
	Delete(*Table, []*Field) string
	Select(*Table, []*Field) string
	SelectCount(*Table, []*Field) string
	SelectRange(*Table, []*Field) string
	SelectByUniqueIndex(t *Table, fields []*Field, index *Index) string
	Param(int) string
	Token(int) string
}

func New(dialect int) Dialect {
	switch dialect {
	case POSTGRES:
		return newPosgres()
	case MYSQL:
		return newMysql()
	default:
		return newSqlite()
	}
}
