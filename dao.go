// Copyright 2016 polaris. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	polaris@studygolang.com

package dbutil

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"github.com/polaris1119/goutils"
)

const (
	dateTimeLayout = "2006-01-02 15:04:05"

	defaultMaxIdleConns = 2
)

var db *sql.DB

// InitDB maxes 第一个用于设置 SetMaxIdleConns，第二个用于设置 SetMaxOpenConns
func InitDB(dsn string, maxes ...int) {
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	num := len(maxes)
	if num > 1 {
		db.SetMaxIdleConns(maxes[0])
	} else {
		db.SetMaxIdleConns(defaultMaxIdleConns)
	}

	if num > 2 {
		db.SetMaxOpenConns(maxes[1])
	}
}

type Dao struct {
	table string

	fields string

	where    string
	whereVal []interface{}

	set    string
	setVal []interface{}

	orderBy       string
	offset, total int

	tx    *sql.Tx
	txErr error
}

func NewDao() *Dao {
	return &Dao{
		whereVal: []interface{}{},
		setVal:   []interface{}{},
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

	stmt, err := d.prepare(d.genFindSql())
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(d.whereVal...)
	if err != nil {
		return err
	}
	defer rows.Close()

	d.reset()

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
	d.fillStructFields(fieldNum, entityType, entityVal, dests, columns)
	return nil
}

// 如果 slice 的 len != cap，entitiesb必须是指向 model slice 类型的指针
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

	if tabler, ok := entitiesVal.Index(0).Interface().(Tabler); ok && d.table == "" {
		d.table = tabler.Table()
	}

	d.fetchStructFieldNames(entityType)

	stmt, err := d.prepare(d.genFindSql())
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(d.whereVal...)
	if err != nil {
		return err
	}
	defer rows.Close()

	d.reset()

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
		d.fillStructFields(fieldNum, entityType, entityVal, dests, columns)
		entitiesVal.Index(colNum).Set(entityVal.Addr())

		colNum++
	}

	return nil
}

func (d *Dao) FindBySql(strSql string, args ...interface{}) (*sql.Rows, error) {
	stmt, err := d.prepare(strSql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (d *Dao) FindOneBySql(strSql string, args ...interface{}) error {
	return nil
}

func (d *Dao) fillStructFields(fieldNum int, entityType reflect.Type, entityVal reflect.Value, dests []interface{}, columns []string) {
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
}

func (d *Dao) Set(set string, setVal ...interface{}) *Dao {
	d.set = set
	d.setVal = setVal
	return d
}

func (d *Dao) Insert(entity interface{}) (int64, error) {

	if creater, ok := entity.(Creater); ok {
		return creater.Create()
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

	result, err := d.exec(strSql)
	if err != nil {
		return 0, err
	}
	d.reset()

	return result.LastInsertId()
}

func (d *Dao) Update() (int64, error) {

	strSql := d.genUpdateSql()

	args := append(d.setVal, d.whereVal...)
	result, err := d.exec(strSql, args...)
	if err != nil {
		return 0, err
	}

	d.reset()

	return result.RowsAffected()
}

func (d *Dao) Persist(entity interface{}, updateField string) (int64, error) {

	if updater, ok := entity.(Updater); ok {
		return updater.Update()
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

	updateFields := strings.Split(updateField, ",")

	// 记录没有设置 pk tag 的 Model 默认 pk id
	var idPk interface{}

	fieldNum := entityVal.NumField()
	for i := 0; i < fieldNum; i++ {
		columnName := d.fetchStructFieldName(entityType, i)

		// 是主键
		if entityType.Field(i).Tag.Get("pk") != "" {
			d.where += " AND " + columnName + "=?"
			d.whereVal = append(d.whereVal, entityVal.Field(i).Interface())
		} else {
			pos := SearchString(updateFields, columnName)
			if pos != -1 {
				d.set += "," + columnName + "=?"
				d.setVal = append(d.setVal, entityVal.Field(i).Interface())
			}

			if columnName == "id" {
				idPk = entityVal.Field(i).Interface()
			}
		}
	}

	// 去除多余的字符
	if d.where != "" {
		d.where = d.where[5:] // 开始的 " AND "
	}
	if d.where == "" {
		d.where = "id=?"
		d.whereVal = []interface{}{idPk}
	}
	if d.set != "" {
		d.set = d.set[1:] // 开始的 ","
	}

	strSql := d.genUpdateSql()
	args := append(d.setVal, d.whereVal...)
	result, err := d.exec(strSql, args...)
	if err != nil {
		return 0, err
	}
	d.reset()

	return result.RowsAffected()
}

func (d *Dao) Delete(entity interface{}) {

}

// Begin 开启事务
func (d *Dao) Begin() {
	d.tx, d.txErr = db.Begin()
	return
}

func (d *Dao) prepare(query string) (*sql.Stmt, error) {
	if d.tx != nil {
		if d.txErr != nil {
			return nil, errors.New("事务开启错误")
		}
		return d.tx.Prepare(query)
	}
	return db.Prepare(query)
}

func (d *Dao) exec(query string, args ...interface{}) (sql.Result, error) {
	if d.tx != nil {
		if d.txErr != nil {
			return nil, errors.New("事务开启错误")
		}
		return d.tx.Exec(query, args...)
	}

	return db.Exec(query, args...)
}

// Commit 提交事务
func (d *Dao) Commit() error {
	return d.tx.Commit()
}

// Rollback 回滚事务
func (d *Dao) Rollback() error {
	return d.tx.Rollback()
}

func (d *Dao) fetchStructFieldNames(entityType reflect.Type) string {
	if d.fields != "" {
		return d.fields
	}

	buffer := goutils.NewBuffer()

	numField := entityType.NumField()
	for i := 0; i < numField; i++ {
		columnName := d.fetchStructFieldName(entityType, i)

		buffer.Append(",").Append(columnName)
	}

	d.fields = buffer.String()[1:]

	return d.fields
}

func (d *Dao) fetchStructFieldName(entityType reflect.Type, i int) string {
	tag := entityType.Field(i).Tag
	columnName := tag.Get("db")
	if columnName == "" {
		columnName = tag.Get("json")
		if columnName == "" {
			columnName = UnderscoreName(entityType.Field(i).Name)
		}
	}
	return columnName
}

func (d *Dao) fetchStructFieldValues(entityVal reflect.Value) string {
	buffer := goutils.NewBuffer()

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
	buffer := goutils.NewBuffer()

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

func (d *Dao) genUpdateSql() string {
	buffer := goutils.NewBuffer()

	buffer.Append(fmt.Sprintf("UPDATE %s SET %s", d.table, d.set))

	if d.where == "" {
		// 为了安全，不允许没有条件更新所有数据
		return ""
	}
	buffer.Append(" WHERE ").Append(d.where)

	return buffer.String()
}

func (d *Dao) reset() {
	d.table = ""
	d.fields = ""

	d.where = ""
	d.whereVal = []interface{}{}

	d.set = ""
	d.setVal = []interface{}{}

	d.orderBy = ""
	d.offset = 0
	d.total = 0
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
	buffer := goutils.NewBuffer()
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
