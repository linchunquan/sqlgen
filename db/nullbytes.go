package db

import (
	"database/sql/driver"
	"fmt"
)

/*type NullBytes struct {
    Bytes  []byte
    Valid bool // 标识数据是否有效（非 NULL）
}

// 实现 database/sql.Scanner 接口
func (n *NullBytes) Scan(value interface{}) error {
    if value == nil {
        n.Bytes, n.Valid = nil, false
        return nil
    }
    n.Valid = true
    n.Bytes = value.([]byte)
    return nil
}*/

type NullBytes struct {
	Bytes []byte
	Valid bool // 标识数据是否有效（非 NULL）
}

func (n *NullBytes) Scan(value interface{}) error {
	if value == nil {
		n.Bytes, n.Valid = nil, false
		return nil
	}

	n.Valid = true

	switch v := value.(type) {
	case []byte:
		// 深拷贝避免引用驱动缓冲区
		n.Bytes = make([]byte, len(v))
		copy(n.Bytes, v)
	case string:
		// 处理字符串类型转换
		n.Bytes = []byte(v)
	default:
		// 明确拒绝不支持的類型
		return fmt.Errorf("NullBytes.Scan: 不支持的類型 %T", value)
	}

	return nil
}

// 可选：实现driver.Valuer接口以支持写入操作
func (n NullBytes) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Bytes, nil
}
