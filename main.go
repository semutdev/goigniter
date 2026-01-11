package main

import (
	"fmt"
	"goigniter/config"
	_ "goigniter/controllers"
	_ "goigniter/controllers/admin"
	"goigniter/database"
	"goigniter/libs"
	"goigniter/models"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Template struct {
	viewsDir string
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Cek apakah template ada di subfolder (misal: auth/login atau admin/product/index)
	parts := strings.Split(name, "/")

	if len(parts) >= 2 {
		// Template dengan layout (subfolder)
		folder := parts[0]
		layoutPath := filepath.Join(t.viewsDir, folder, "layout.html")
		var pagePath string

		if len(parts) == 2 {
			// auth/login -> views/auth/login.html
			pagePath = filepath.Join(t.viewsDir, folder, parts[1]+".html")
		} else if len(parts) == 3 {
			// admin/product/index -> views/admin/product/index.html
			pagePath = filepath.Join(t.viewsDir, folder, parts[1], parts[2]+".html")
		}

		// Cek apakah layout exists
		if _, err := os.Stat(layoutPath); err == nil {
			// Parse layout + page bersama dengan FuncMap
			tmpl, err := template.New("").Funcs(libs.TemplateFuncs()).ParseFiles(layoutPath, pagePath)
			if err != nil {
				return err
			}
			return tmpl.ExecuteTemplate(w, name, data)
		}
	}

	// Fallback: parse semua template dengan FuncMap
	tmpl := parseTemplates(t.viewsDir)
	return tmpl.ExecuteTemplate(w, name, data)
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// parseTemplates parse semua template termasuk subfolder
func parseTemplates(viewsDir string) *template.Template {
	tmpl := template.New("").Funcs(libs.TemplateFuncs())

	err := filepath.Walk(viewsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directory
		if info.IsDir() {
			return nil
		}

		// Hanya parse file .html
		if !strings.HasSuffix(path, ".html") {
			return nil
		}

		// Baca file
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Parse template
		_, err = tmpl.Parse(string(content))
		return err
	})

	if err != nil {
		panic("Error parsing templates: " + err.Error())
	}

	return tmpl
}

func main() {
	// load env
	godotenv.Load()

	// connect to DB
	config.ConnectDB()

	// auto migrate tables
	config.DB.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.LoginAttempt{},
		&models.Product{},
	)

	// run seeder jika DB_SEED=true
	if os.Getenv("DB_SEED") == "true" {
		database.Seed(config.DB)
	}

	e := echo.New()
	e.Debug = true

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogMethod: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fmt.Printf("[%s] %s status=%d\n", v.Method, v.URI, v.Status)
			return nil
		},
	}))

	e.Validator = &CustomValidator{validator: validator.New()}

	// setup static
	e.Use(middleware.Recover())
	e.Static("/static", "public")

	// Setup template renderer dengan dukungan multi-layout
	t := &Template{
		viewsDir: "views",
	}
	e.Renderer = t

	libs.AutoRoute(e, libs.GetRegistry())

	// default route
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusTemporaryRedirect, "/welcome")
	})

	port := os.Getenv("APP_PORT")

	if port == "" {
		port = ":6789"
	}

	e.Logger.Fatal(e.Start(port))
}
