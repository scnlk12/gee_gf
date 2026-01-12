package dialect

import "reflect"

// ORM框架往往需要兼容多种数据库，不同数据库支持的数据类型是有差异的，即使功能相同，在sql语句的表达上也可能有差异
// 需要将这部分差异提取出来，每一种数据库分别实现，实现最大程度的复用和解耦
var dialectMap = map[string]Dialect{}

type Dialect interface {
	// 将go语言的类型转换为当前数据库的数据类型
	DataTypeOf(typ reflect.Value) string
	// 返回某个表是否存在的SQL语句 参数是表名tableName
	TableExistSQL(tableName string) (string, []interface{})
}

// 注册dialect实例 
// 如果新增加对某个数据库的支持 调用RegisterDialect即可注册到全局
func RegisterDialect(name string, dialect Dialect) {
	dialectMap[name] = dialect
}

// 获取dialect实例
func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectMap[name]
	return
}