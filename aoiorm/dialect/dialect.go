package dialect

import (
	"AoiFramework/aoiorm/olog"
	"fmt"
	"reflect"
)

//用户自己注册map
var dialectMap = map[string]Dialect{}

// Dialect 各个数据库实现的转化sql接口
type Dialect interface {
	// DtaTypeof go类型转数据库，不同数据库定制不同
	DtaTypeof(typ reflect.Value) string
	// TableExistSQL 传入表名，返回某个表是否存在
	TableExistSQL(tableName string) (string, []interface{})
}

func RegisterDialect(name string, dialect Dialect) error {
	if _, ok := dialectMap[name]; ok {
		err := fmt.Errorf("dup dialect")
		olog.Error(err)
		return err
	}
	dialectMap[name] = dialect
	return nil
}

func GetDialect(name string) (Dialect, bool) {
	dia, ok := dialectMap[name]
	return dia, ok
}
