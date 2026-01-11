package main

import (
	"fmt"
	"goigniter/config"
	_ "goigniter/controllers" // auto-register semua controller via init()
	"goigniter/libs"
	"goigniter/models"
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	// load env
	godotenv.Load()

	// connect to DB
	config.ConnectDB()

	// auto create table
	config.DB.AutoMigrate(&models.User{})

	e := echo.New()
	e.Debug = true

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogMethod: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			// Ini akan mencetak log ke terminal: "GET / 200"
			fmt.Printf("[%s] %s status=%d\n", v.Method, v.URI, v.Status)
			return nil
		},
	}))

	e.Validator = &CustomValidator{validator: validator.New()}

	// setup static
	e.Use(middleware.Recover())
	e.Static("/static", "public")

	t := &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
	e.Renderer = t

	// e.GET("/", controllers.Index)
	// e.POST("/users/add", controllers.Store)
	// e.DELETE("/users/delete/:id", controllers.Delete)
	// e.GET("/users/edit/:id", controllers.EditForm)
	// e.GET("/users/row/:id", controllers.GetRow)
	// e.PUT("/users/update/:id", controllers.Update)

	// Auto-route: semua controller sudah ter-register via init()
	libs.AutoRoute(e, libs.GetRegistry())

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusTemporaryRedirect, "/users/index")
	})

	port := os.Getenv("APP_PORT")

	if port == "" {
		port = ":6789"
	}

	e.Logger.Fatal(e.Start(port))
}
