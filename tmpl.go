package main

// template to create a constant variable.
var sConst = `
const %s = %s
`

// template to wrap a string in multi-line quotes.
var sQuote = "`\n%s\n`"

// template to declare the package name.
var sPackage = `
package %s

// THIS FILE WAS AUTO-GENERATED. DO NOT MODIFY.
`

// template to delcare the package imports.
var sImport = `
import (
	%s
)
`

// function template to scan a single row.
const sScanRow = `
func scan%s(row *sql.Row) (*%s, error) {
	%s

	err := row.Scan(
		%s
	)
	if err != nil {
		return nil, err
	}

	v := &%s{}
	%s

	return v, nil
}
`

// function template to scan multiple rows.
const sScanRows = `
func scan%s(rows *sql.Rows) ([]*%s, error) {
	var err error
	var vv []*%s

	%s
	for rows.Next() {
		err = rows.Scan(
			%s
		)
		if err != nil {
			return vv, err
		}

		v := &%s{}
		%s
		vv = append(vv, v)
	}
	return vv, rows.Err()
}
`

const sSliceRow = `
func slice%s(v *%s) []interface{} {
	%s
	%s

	return []interface{}{
		%s
	}
}
`

const sGenericSelectRow = `
func genericSelect%s(db db.SimpleDB, query string, args ...interface{}) (*%s, error) {
	row := db.QueryRow(query, args...)
	return scan%s(row)
}
`

// function template to select multiple rows.
const sGenericSelectRows = `
func genericSelect%s(db db.SimpleDB, query string, args ...interface{}) ([]*%s, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scan%s(rows)
}
`

// function template to insert a single row.
const sGenericInsert = `
func genericInsert%s(db db.SimpleDB, query string, v *%s) error {

	res, err := db.Exec(query, slice%s(v)[1:]...)
	if err != nil {
		return err
	}

	v.ID, err = res.LastInsertId()
	return err
}
`

// function template to update a single row.
const sGenericUpdate = `
func genericUpdate%s(db db.SimpleDB, query string, v *%s) error {

	args := slice%s(v)[1:]
	args = append(args, v.ID)
	_, err := db.Exec(query, args...)
	return err 
}
`

const sInsert = `
func Insert%s(db db.SimpleDB,  v *%s) error {

	res, err := db.Exec(%s, slice%s(v)[1:]...)
	if err != nil {
		return err
	}

	v.ID, err = res.LastInsertId()
	return err
}
`
const sDelete = `
func Delete%s%s(db db.SimpleDB, %s) error {
	args := []interface{}{%s}
	_, err := db.Exec(%s, args...)
	return err
}
`

const sUpdate = `
func Update%s%s(db db.SimpleDB, v *%s) error {
	args := slice%s(v)
    args = append(args,%s)
	_, err := db.Exec(%s, args...)
	return err
}
`

const sGetBy = `
func Get%s%s(db db.SimpleDB, %s) (*%s, error) {
	args := []interface{}{%s}
	v, err :=  genericSelect%s(db, %s, args...)
	return v, err
}
`

const sFindByIndex = `
func Find%ss%s(db db.SimpleDB, %s) ([]*%s, error) {
	args := []interface{}{%s}
	v, err :=  genericSelect%ss(db, %s, args...)
	return v, err
}
`

const sFindByIndexInRange = `
func Find%ss%sInRange(db db.SimpleDB, %s, limit int64, offset int64) ([]*%s, error) {
	args := []interface{}{%s, limit, offset}
	v, err :=  genericSelect%ss(db, %s, args...)
	return v, err
}
`

const sFindByForeignKey = `
func Find%ssOf%s%s(db db.SimpleDB, %s) ([]*%s, error) {
	args := []interface{}{%s}
	v, err :=  genericSelect%ss(db, %s, args...)
	return v, err
}
`

const sFindByForeignKeyInRange = `
func Find%ssOf%s%sInRange(db db.SimpleDB, %s, limit int64, offset int64) ([]*%s, error) {
	args := []interface{}{%s, limit, offset}
	v, err :=  genericSelect%ss(db, %s, args...)
	return v, err
}
`

const sGetByForeignKey = `
func Get%sOf%s%s(db db.SimpleDB, %s) (*%s, error) {
	args := []interface{}{%s}
	v, err :=  genericSelect%s(db, %s, args...)
	return v, err
}
`

const sFindAll = `
func FindAll%ss(db db.SimpleDB) ([]*%s, error) {
	args := []interface{}{}
	v, err :=  genericSelect%ss(db, %s, args...)
	return v, err
}
`

const sFindAllInRange = `
func FindAll%ssInRange(db db.SimpleDB, limit int64, offset int64) ([]*%s, error) {
	args := []interface{}{limit, offset}
	v, err :=  genericSelect%ss(db, %s, args...)
	return v, err
}
`

const sCount = `
func Count%s(db db.SimpleDB)(int, error){
    var count int
	row := db.QueryRow(%s)
	err := row.Scan(&count)
	return count, err
}
`

const sCountByIndex = `
func Count%s%s(db db.SimpleDB, %s)(int, error){
    var count int
    args := []interface{}{%s}
	row := db.QueryRow(%s, args...)
	err := row.Scan(&count)
	return count, err
}
`
