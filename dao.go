package dbutil

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
)

const (
	dateTimeLayout = "2006-01-02 15:04:05"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "root:@tcp(localhost:3306)/studygolang?charset=utf8")
	if err != nil {
		panic(err)
	}
}

type Dao struct {
	table         string
	fields        string
	where         string
	whereVal      []interface{}
	orderBy       string
	offset, total int
}

func NewDao() *Dao {
	return &Dao{
	//fields: "*",
	}
}

func (d *Dao) Table(table string) *Dao {
	d.table = table
	return d
}

func (d *Dao) Fileds(fields string) *Dao {
	d.fields = fields
	return d
}

func (d *Dao) Where(condition string, args ...interface{}) *Dao {
	d.where = condition
	d.whereVal = args
	return d
}

func (d *Dao) OrderBy(orderBy string) *Dao {
	d.orderBy = orderBy
	return d
}

func (d *Dao) Limit(total int, offset ...int) *Dao {
	d.total = total
	if len(offset) > 1 {
		d.offset = offset[0]
	}
	return d
}

func (d *Dao) FindOne(entity interface{}) error {
	if reader, ok := entity.(Reader); ok {
		return reader.Read()
	}

	entityType := reflect.TypeOf(entity)
	if entityType.Kind() != reflect.Ptr && entityType.Elem().Kind() != reflect.Struct {
		return errors.New("entity must the pointer of struct")
	}

	entityType = entityType.Elem()
	entityVal := reflect.ValueOf(entity).Elem()

	if tabler, ok := entity.(Tabler); ok && d.table == "" {
		d.table = tabler.Table()
	}

	d.fetchStructFieldNames(entityType)

	stmt, err := db.Prepare(d.genFindSql())
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(d.whereVal...)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	dests := make([]interface{}, len(columns))
	if rows.Next() {
		for i := range dests {
			var dest interface{}
			dests[i] = &dest
		}
		err = rows.Scan(dests...)
		if err != nil {
			return err
		}
	}

	fieldNum := entityVal.NumField()
	for i := 0; i < fieldNum; i++ {
		structField := entityType.Field(i)
		columnName := structField.Tag.Get("db")
		if columnName == "" {
			columnName = structField.Tag.Get("json")
			if columnName == "" {
				columnName = UnderscoreName(structField.Name)
			}
		}

		pos := SearchString(columns, columnName)
		destVal := dests[pos]

		filedVal := entityVal.Field(i)
		if filedVal.CanSet() {
			assignTo(filedVal, reflect.ValueOf(destVal).Elem())
		}
	}
	return nil
}

// entities 是指向 model slice 类型的指针
func (d *Dao) FindAll(entities interface{}) error {
	entitiesVal := reflect.ValueOf(entities)
	entitiesType := reflect.TypeOf(entities)
	if entitiesType.Kind() != reflect.Ptr {
		if entitiesVal.Len() != entitiesVal.Cap() {
			return errors.New("len!=cap, so entities must be pointer of slice")
		}
	} else {
		entitiesType = entitiesType.Elem()
		entitiesVal = entitiesVal.Elem()
		// 避免 append 产生新的 slice
		entitiesVal.SetLen(entitiesVal.Cap())
	}

	if entitiesType.Kind() != reflect.Slice {
		return errors.New("entities must be slice")
	}

	entityType := entitiesType.Elem()
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	if entityType.Kind() != reflect.Struct {
		return errors.New("the element of slice(entities) must be struct or pointer of struct")
	}

	fieldNum := entityType.NumField()

	d.fetchStructFieldNames(entityType)

	stmt, err := db.Prepare(d.genFindSql())
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(d.whereVal...)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	dests := make([]interface{}, len(columns))
	for i := range dests {
		var dest interface{}
		dests[i] = &dest
	}

	var (
		colNum    = 0
		entityVal reflect.Value
	)
	for rows.Next() {
		err = rows.Scan(dests...)
		if err != nil {
			return err
		}

		entityVal = reflect.New(entityType).Elem()

		for i := 0; i < fieldNum; i++ {
			structField := entityType.Field(i)
			columnName := structField.Tag.Get("db")
			if columnName == "" {
				columnName = structField.Tag.Get("json")
				if columnName == "" {
					columnName = UnderscoreName(structField.Name)
				}
			}

			pos := SearchString(columns, columnName)
			destVal := dests[pos]

			filedVal := entityVal.Field(i)
			if filedVal.CanSet() {
				assignTo(filedVal, reflect.ValueOf(destVal).Elem())
			}
		}
		entitiesVal.Index(colNum).Set(entityVal.Addr())

		colNum++
	}

	return nil
}

func (d *Dao) Insert(entity interface{}) (int64, error) {

	if cruder, ok := entity.(Creater); ok {
		return cruder.Create()
	}

	entityType := reflect.TypeOf(entity)
	entityVal := reflect.ValueOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
		entityVal = entityVal.Elem()
	}

	if tabler, ok := entity.(Tabler); ok {
		d.table = tabler.Table()
	}

	strSql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", d.table, d.fetchStructFieldNames(entityType), d.fetchStructFieldValues(entityVal))

	fmt.Println(strSql)
	result, err := db.Exec(strSql)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (d *Dao) fetchStructFieldNames(entityType reflect.Type) string {
	if d.fields != "" {
		return d.fields
	}

	buffer := NewBuffer()

	numField := entityType.NumField()
	for i := 0; i < numField; i++ {
		tag := entityType.Field(i).Tag
		columnName := tag.Get("db")
		if columnName == "" {
			columnName = tag.Get("json")
			if columnName != "" {
				columnName = UnderscoreName(entityType.Field(i).Name)
			}
		}

		buffer.Append(",").Append(columnName)
	}

	d.fields = buffer.String()[1:]

	return d.fields
}

func (d *Dao) fetchStructFieldValues(entityVal reflect.Value) string {
	buffer := NewBuffer()

	numField := entityVal.NumField()
	for i := 0; i < numField; i++ {
		buffer.Append(",")

		fieldVal := entityVal.Field(i)
		switch fieldVal.Kind() {
		case reflect.String:
			buffer.Append("'").Append(fieldVal.String()).Append("'")
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
			buffer.AppendUint(fieldVal.Uint())
		}
	}

	return buffer.String()[1:]
}

func (d *Dao) genFindSql() string {
	buffer := NewBuffer()

	buffer.Append(fmt.Sprintf("SELECT %s FROM %s", d.fields, d.table))

	if d.where != "" {
		buffer.Append(" WHERE ").Append(d.where)
	}

	if d.orderBy != "" {
		buffer.Append(" ORDER BY ").Append(d.orderBy)
	}

	if d.total > 0 {
		buffer.Append(fmt.Sprintf(" LIMIT %d OFFSET %d", d.total, d.offset))
	}

	return buffer.String()
}

type Tabler interface {
	Table() string
}

type Creater interface {
	Create() (int64, error)
}

type Reader interface {
	Read() error
}

type Updater interface {
	Update() (int64, error)
}

type Deleter interface {
	Delete() (int64, error)
}

type Cruder interface {
	Creater
	Reader
	Updater
	Deleter
}

// 驼峰式写法转为下划线写法
func UnderscoreName(name string) string {
	buffer := NewBuffer()
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer.AppendRune('_')
			}
			buffer.AppendRune(unicode.ToLower(r))
		} else {
			buffer.AppendRune(r)
		}
	}

	return buffer.String()
}

func SearchString(slice []string, s string) int {
	for i, v := range slice {
		if s == v {
			return i
		}
	}

	return -1
}

func assignTo(target, value reflect.Value) {
	i := value.Interface()
	switch v := i.(type) {
	case uint8, uint16, uint, uint32, uint64:
		switch target.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int32, reflect.Int64:
			target.SetInt(int64(v.(uint64)))
		default:
			target.SetUint(v.(uint64))
		}

	case int8, int16, int, int32, int64:
		switch target.Kind() {
		case reflect.Uint8, reflect.Uint16, reflect.Uint, reflect.Uint32, reflect.Uint64:
			target.SetUint(uint64(v.(int64)))
		case reflect.Bool:
			if v.(int64) > 0 {
				target.SetBool(true)
			} else {
				target.SetBool(false)
			}
		default:
			target.SetInt(v.(int64))
		}

	case string:
		target.SetString(v)

	case []byte:
		kind := target.Kind()
		if kind == reflect.String {
			target.SetString(string(v))
		} else if kind == reflect.Struct {
			// time.Time
			dt, err := time.ParseInLocation(dateTimeLayout, string(v), time.Local)
			if err != nil {
				// TODO:
				break
			}
			target.Set(reflect.ValueOf(dt))
		} else {
			target.SetBytes(v)
		}

	case float32:
		target.SetFloat(float64(v))
	case float64:
		target.SetFloat(v)
	case bool:
		target.SetBool(v)
	}
}
