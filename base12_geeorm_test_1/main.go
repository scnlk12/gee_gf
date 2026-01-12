package main

import (
	"fmt"
	"geeorm"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	engine, _ := geeorm.NewEngine("sqlite3", "gee.db")
	defer engine.Close()
	s := engine.NewSession()
	_, _ = s.Raw("drop table if exists User;").Exec()
	_, _ = s.Raw("create table User(Name text)").Exec()
	_, _ = s.Raw("create table User(Name text)").Exec()
	result, _ := s.Raw("insert into User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	count, _ := result.RowsAffected()
	fmt.Println("Exec success, %d affected\n", count)
}
