package schema

import (
	"geeorm/dialect"
	"go/ast"
	"reflect"
)

// Field represents a column of database
type Field struct {
	Name string
	Type string
	Tag  string
}

// Schema represents a table of database
// 包含被映射的对象Model 表名Name 字段Fields
// FieldNames包含所有字段名/列名 fieldMap记录字段名和Field的映射关系
type Schema struct {
	Model      interface{}
	Name       string
	Fields     []*Field
	FieldNames []string
	fieldMap   map[string]*Field
}

func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

func Parse(dest interface{}, d dialect.Dialect) *Schema {
	// reflect.Indirect()获取指针指向的实例
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(), // 获取结构体的名称作为表名
		fieldMap: make(map[string]*Field),
	}

	// NumField() 获取实例的字段的个数 通过下标获取到指定字段
	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,                                              // 字段名
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))), // 字段类型 转换为数据库的字段类型
			}
			if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}
