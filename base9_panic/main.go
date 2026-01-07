package main

import (
	"gee"
	"net/http"
)

func main() {
	r := gee.Default()

	r.GET("/", func(ctx *gee.Context) {
		ctx.String(http.StatusOK, "Hello Geektutu\n")
	})
	// index out of range for testing Recovery()
	r.GET("/panic", func(ctx *gee.Context) {
		names := []string{"geektutu"}
		ctx.String(http.StatusOK, names[100])
	})

	r.Run(":9999")
}