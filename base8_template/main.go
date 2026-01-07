package main

import (
	"fmt"
	"gee"
	"net/http"
	"text/template"
	"time"
)

type student struct {
	Name string
	Age int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	r := gee.New()
	// 全局中间件
	r.Use(gee.Logger())
	// 自定义模板渲染函数
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})
	r.LoadHTMLGlob("C:/Users/12062/Downloads/gee_gf/gee/static/templates/*")
	r.Static("/assets", "./static")

	stu1 := &student{Name: "Geektutu", Age: 20}
	stu2 := &student{Name: "Jack", Age: 22}
	r.GET("/", func(ctx *gee.Context) {
		ctx.HTML(http.StatusOK, "css.tmpl", nil)
	})
	r.GET("/students", func(ctx *gee.Context) {
		ctx.HTML(http.StatusOK, "arr.tmpl", gee.H{
			"title": "gee",
			"stuArr": [2]*student{stu1, stu2},
		})
	})

	r.GET("/date", func(ctx *gee.Context) {
		ctx.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
			"title": "gee",
			"now": time.Date(2026, 1, 7, 11, 38, 12, 0, time.UTC),
		})
	})

	r.Run(":9999")
}