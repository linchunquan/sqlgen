package db
type NullBytes struct {
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
}
