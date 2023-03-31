package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", hello)

	e.Logger.Fatal(e.Start(":1323"))

	// fmt.Println("this is a test")
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello From Echo library!")
}
