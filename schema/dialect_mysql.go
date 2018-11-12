package schema

import (
	"fmt"
	"log"
)

type mysql struct {
	base
}

func newMysql() Dialect {
	d := &mysql{}
	d.base.Dialect = d
	return d
}

func (d *mysql) Column(f *Field) (_ string) {
	switch f.Type {
	case INTEGER:
		return "INTEGER"
	case LONG:
		return "BIGINT"
	case BOOLEAN:
		return "BOOLEAN"
	case BLOB:
		return "MEDIUMBLOB"
	case VARCHAR:
		// assigns an arbitrary size if
		// none is provided.
		size := f.Size
		if size == 0 {
			size = 512
		}
		return fmt.Sprintf("VARCHAR(%d)", size)
	default:
		return
	}
}

func (d *mysql) Token(v int) (_ string) {
	switch v {
	case AUTO_INCREMENT:
		return "AUTO_INCREMENT"
	case PRIMARY_KEY:
		return "PRIMARY KEY"
	default:
		return
	}
}

// Index returns a SQL statement to create the index.
func (b *mysql) Index(table *Table, index *Index) string {
	log.Printf("create index:%+v", index)
	var obj = "INDEX"
	if index.Unique {
		obj = "UNIQUE INDEX"
	}
	return fmt.Sprintf("CREATE %s %s ON %s (%s);", obj, index.Name, table.Name, b.columns(nil, index.Fields, true, false, false))
}
