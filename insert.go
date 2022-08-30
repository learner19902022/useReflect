package homework

import (
	"errors"
	"reflect"
	"strings"
)

var errInvalidEntity = errors.New("invalid entity")
var errNilEntity = errors.New("entity cannot be nil")
var errEmptyEntity = errors.New("entity cannot be empty")

// below is not good, global should only be used, but not changed
var querySQL strings.Builder
var totalFieldCnt int
var fieldNames map[string]int

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
	//args := make([]interface{}, typ.NumField())
	querySQL = strings.Builder{}
	querySQL.WriteString("INSERT INTO `" + typ.Name() + "`(")
	var args []interface{}
	errMsg := errors.New("")

	// 遍历所有的字段，构造出来的是 INSERT INTO XXX(col1, col2, col3)
	// 在这个遍历的过程中，你就可以把参数构造出来
	// 如果你打算支持组合，那么这里你要深入解析每一个组合的结构体
	// 并且层层深入进去
	totalFieldCnt = 0
	fieldNames = map[string]int{}
	args, errMsg = iterSubStruct(typ, val, args, errMsg)
	//log.Printf("%s", querySQL.String())
	//Make this into a function and do recursion

	// 拼接 VALUES，达成 INSERT INTO XXX(col1, col2, col3) VALUES
	// 再一次遍历所有的字段，要拼接成 INSERT INTO XXX(col1, col2, col3) VALUES(?,?,?)
	// 注意，在第一次遍历的时候我们就已经拿到了参数的值，所以这里就是简单拼接 ?,?,?
	querySQL.WriteString(") VALUES(")
	for iterField := 0; iterField < totalFieldCnt; iterField++ { //typ.NumField()
		if iterField != 0 {
			querySQL.WriteString(",")
		}
		querySQL.WriteString("?")
	}
	querySQL.WriteString(");")

	//log.Printf("%s", querySQL.String())
	//panic("implement me")
	return querySQL.String(), args, nil
}

func iterSubStruct(typ reflect.Type, val reflect.Value, args []interface{}, errMsg error) ([]interface{}, error) {
	if typ.NumField() == 0 {
		querySQL = strings.Builder{}
		return nil, errEmptyEntity
	}
	//log.Printf("%s", "here is ok")
	for iterField := 0; iterField < typ.NumField(); iterField++ {
		fd := typ.Field(iterField)
		fdTyp := fd.Type
		fdVal := val.Field(iterField)
		//log.Printf("%s", fdTyp.String())        //see difference here (Type can be the name of struct)
		//log.Printf("%s", fdTyp.Kind().String()) //see difference here (Kind is the basic types of go, like struct, interface, etc, if there is a type StructName struct, Type will give StructName while Kind will give "struct")
		if fdTyp.String() == "sql.NullString" || fdTyp.Kind() != reflect.Struct {
			if fieldNames[fd.Name] == 1 {
				continue
			}
			fieldNames[fd.Name] = 1
			totalFieldCnt += 1
			args = append(args, fdVal.Interface())
			if iterField != 0 {
				querySQL.WriteString(",")
			}
			querySQL.WriteString("`" + typ.Field(iterField).Name + "`")
		} else {
			if fieldNames[fd.Name] == 1 {
				continue
			}
			fieldNames[fd.Name] = 1
			//typSubStruct := fdTyp //reflect.TypeOf(fdTyp)
			//log.Printf("%v", val.Field(iterField))
			//valSubStruct := val.Field(iterField)                            //reflect.ValueOf(val.Field(iterField))
			args, errMsg = iterSubStruct(fdTyp, fdVal, args, errMsg) //typSubStruct, valSubStruct, args, errMsg)
		}
	}
	errMsg = nil
	//log.Printf("%s", querySQL.String())
	return args, errMsg
}

//why the input is interface, but it is ok to transmit struct as input??
