package geeorm

import "testing"

type User struct {
	Name string `geeorm:"primary key"`
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	engine, _ := NewEngine("sqlite3", "gee.db")
	defer engine.Close()
	s := engine.NewSession().Model(&User{})
	_ = s.DropTable()
	_ = s.CreateTable()
	if !s.HasTable() {
		t.Fatal("Failed to create table User")
	}
}