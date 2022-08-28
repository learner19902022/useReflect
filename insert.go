package homework

import (
	"errors"
	"reflect"
	"strings"
)

var errInvalidEntity = errors.New("invalid entity")
var errNilEntity = errors.New("entity cannot be nil")
var errEmptyEntity = errors.New("entity cannot be empty")

// InsertStmt 作业里面我们这个只是生成 SQL，所以在处理 sql.NullString 之类的接口
// 只需要判断有没有实现 driver.Valuer 就可以了
func InsertStmt(entity interface{}) (string, []interface{}, error) {
	// val := reflect.ValueOf(entity)
	// typ := val.Type()
	// 检测 entity 是否符合我们的要求
	// 我们只支持有限的几种输入

	if entity == nil {
		return "", nil, errNilEntity
	}
	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return "", nil, errInvalidEntity
	}
	//val.IsNil()
	if typ.NumField() == 0 {
		return "", nil, errEmptyEntity
	}

	// 使用 strings.Builder 来拼接 字符串
	// bd := strings.Builder{}
	// 构造 INSERT INTO XXX，XXX 是你的表名，这里我们直接用结构体名字
	querySQL := strings.Builder{}
	querySQL.WriteString("INSERT INTO `" + typ.Name() + "`(")

	// 遍历所有的字段，构造出来的是 INSERT INTO XXX(col1, col2, col3)
	// 在这个遍历的过程中，你就可以把参数构造出来
	// 如果你打算支持组合，那么这里你要深入解析每一个组合的结构体
	// 并且层层深入进去
	args := make([]interface{}, typ.NumField())
	for iterField := 0; iterField < typ.NumField(); iterField++ {
		args[iterField] = val.Field(iterField).Interface() //typ.Field(iterField))
		if iterField != 0 {
			querySQL.WriteString(",")
		}
		querySQL.WriteString("`" + typ.Field(iterField).Name + "`")
	}

	// 拼接 VALUES，达成 INSERT INTO XXX(col1, col2, col3) VALUES
	// 再一次遍历所有的字段，要拼接成 INSERT INTO XXX(col1, col2, col3) VALUES(?,?,?)
	// 注意，在第一次遍历的时候我们就已经拿到了参数的值，所以这里就是简单拼接 ?,?,?
	querySQL.WriteString(") VALUES(")
	for iterField := 0; iterField < typ.NumField(); iterField++ {
		if iterField != 0 {
			querySQL.WriteString(",")
		}
		querySQL.WriteString("?")
	}
	querySQL.WriteString(");")

	//panic("implement me")
	return querySQL.String(), args, nil
}

//why the input is interface, but it is ok to transmit struct as input??
