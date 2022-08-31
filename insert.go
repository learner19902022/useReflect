package homework

import (
	"errors"
	"reflect"
	"strings"
)

var errInvalidEntity = errors.New("invalid entity")
var errNilEntity = errors.New("entity cannot be nil")
var errEmptyEntity = errors.New("entity cannot be empty")

// strings.Builder不允许以函数参数形式传递或复制，因此选用了全局变量
// 为了简化递归函数参数传递，使用"封装"的全局变量接口来实现查询重复field和统计总的field数
var querySQL strings.Builder
var markRed markRedundant

// InsertStmt 作业里面我们这个只是生成 SQL，所以在处理 sql.NullString 之类的接口
// 只需要判断有没有实现 driver.Valuer 就可以了

// InsertStmt 为什么输入的是interface，但是传结构体也可以？结构体也是一种接口么？
func InsertStmt(entity interface{}) (string, []interface{}, error) {
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
	if typ.NumField() == 0 {
		return "", nil, errEmptyEntity
	}

	// 使用 strings.Builder 来拼接 字符串
	// bd := strings.Builder{}
	// 构造 INSERT INTO XXX，XXX 是你的表名，这里我们直接用结构体名字
	querySQL = strings.Builder{}
	querySQL.WriteString("INSERT INTO `" + typ.Name() + "`(")
	var args []interface{}
	errMsg := errors.New("")
	markRed = markRed.init()

	// 遍历所有的字段，构造出来的是 INSERT INTO XXX(col1, col2, col3)
	// 在这个遍历的过程中，你就可以把参数构造出来
	// 如果你打算支持组合，那么这里你要深入解析每一个组合的结构体
	// 并且层层深入进去

	//使用递归函数拆解组合中的结构体，具体逻辑参见后文函数体
	args, errMsg = iterSubStruct(typ, val, args, errMsg)

	// 拼接 VALUES，达成 INSERT INTO XXX(col1, col2, col3) VALUES
	// 再一次遍历所有的字段，要拼接成 INSERT INTO XXX(col1, col2, col3) VALUES(?,?,?)
	// 注意，在第一次遍历的时候我们就已经拿到了参数的值，所以这里就是简单拼接 ?,?,?
	querySQL.WriteString(") VALUES(")
	for iterField := 0; iterField < markRed.getCnt(); iterField++ { //typ.NumField()
		if iterField != 0 {
			querySQL.WriteString(",")
		}
		querySQL.WriteString("?")
	}
	querySQL.WriteString(");")

	//panic("implement me")
	return querySQL.String(), args, nil
}

func iterSubStruct(typ reflect.Type, val reflect.Value, args []interface{}, errMsg error) ([]interface{}, error) {
	if typ.NumField() == 0 {
		querySQL = strings.Builder{}
		return nil, errEmptyEntity
	}
	for iterField := 0; iterField < typ.NumField(); iterField++ {
		fd := typ.Field(iterField)
		fdTyp := fd.Type
		fdVal := val.Field(iterField)
		//log.Printf("%s", fdTyp.String())        //see difference here (Type can be the name of struct)
		//log.Printf("%s", fdTyp.Kind().String()) //see difference here (Kind is the basic types of go, like struct, interface, etc, if there is a type StructName struct, Type will give StructName while Kind will give "struct")
		if fdTyp.String() == "sql.NullString" || fdTyp.Kind() != reflect.Struct || typ.NumField() == 1 {
			if markRed.isFieldNameDuplicate(fd.Name) {
				continue
			}
			markRed.recordFieldName(fd.Name)
			markRed = markRed.incrementCnt()
			args = append(args, fdVal.Interface())
			if iterField != 0 {
				querySQL.WriteString(",")
			}
			querySQL.WriteString("`" + typ.Field(iterField).Name + "`")
		} else {
			if markRed.isFieldNameDuplicate(fd.Name) {
				continue
			}
			markRed.recordFieldName(fd.Name)
			args, errMsg = iterSubStruct(fdTyp, fdVal, args, errMsg)
		}
	}
	errMsg = nil
	return args, errMsg
}

// 封装全局变量，使用接口访问和改变值。问题是：如何确保同一个package下其他.go文件无法访问全局私有变量？
type markRedundant struct {
	totalFieldCnt int
	fieldNames    map[string]int
}

func (m markRedundant) init() markRedundant {
	m.totalFieldCnt = 0
	m.fieldNames = make(map[string]int, 1)
	return m //why we need return m here, if not, fieldNames is a map without initialization? but why?
}

func (m markRedundant) getCnt() int {
	return m.totalFieldCnt
}

func (m markRedundant) incrementCnt() markRedundant {
	m.totalFieldCnt += 1
	return m //why we need to return m to feedback the changes made to m's field? because it is an int?
}

func (m markRedundant) isFieldNameDuplicate(fdName string) bool {
	return m.fieldNames[fdName] == 1
}

func (m markRedundant) recordFieldName(fdName string) {
	m.fieldNames[fdName] = 1
	//why we do not need to return m in this case???
}
