package admin

import (
	"full-crud/application/libs"
	"full-crud/application/models"
	"net/http"
	"strconv"
	"time"

	"github.com/semutdev/goigniter/system/core"
	"github.com/semutdev/goigniter/system/libraries/database"
)

func init() {
	core.Register(&Product{}, "admin")
}

type Product struct {
	core.Controller
}

// Routes defines custom routes for Product controller
func (p *Product) Routes() map[string]string {
	return map[string]string{
		"Edit":   "edit/:id",
		"Update": "update/:id",
		"Delete": "delete/:id",
	}
}

type ProductForm struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// DataTablesResponse format response for DataTables
type DataTablesResponse struct {
	Draw            int              `json:"draw"`
	RecordsTotal    int64            `json:"recordsTotal"`
	RecordsFiltered int64            `json:"recordsFiltered"`
	Data            []models.Product `json:"data"`
}

// Index displays the product list page
func (p *Product) Index() {
	if !libs.IsLoggedIn(p.Ctx) {
		p.Ctx.Redirect(http.StatusSeeOther, "/auth/login")
		return
	}

	data := core.Map{
		"Title":   "Product Management",
		"Success": libs.GetFlash(p.Ctx, "success"),
		"Error":   libs.GetFlash(p.Ctx, "error"),
	}

	p.Ctx.View("admin/inc/header", data)
	p.Ctx.View("admin/product/index", data)
	p.Ctx.View("admin/inc/footer", data)
}

// Data returns JSON for DataTables
func (p *Product) Data() {
	if !libs.IsLoggedIn(p.Ctx) {
		p.Ctx.JSON(http.StatusUnauthorized, core.Map{"error": "Unauthorized"})
		return
	}

	draw, _ := strconv.Atoi(p.Ctx.Query("draw"))
	start, _ := strconv.Atoi(p.Ctx.Query("start"))
	length, _ := strconv.Atoi(p.Ctx.Query("length"))
	search := p.Ctx.Query("search[value]")
	orderCol := p.Ctx.Query("order[0][column]")
	orderDir := p.Ctx.Query("order[0][dir]")

	if length <= 0 {
		length = 10
	}
	if orderDir == "" {
		orderDir = "asc"
	}

	columns := map[string]string{
		"0": "id",
		"1": "name",
		"2": "price",
		"3": "stock",
		"4": "created_at",
	}

	orderColumn := columns[orderCol]
	if orderColumn == "" {
		orderColumn = "id"
	}

	var products []models.Product
	var totalRecords int64
	var filteredRecords int64

	// Total records
	totalRecords, _ = database.Table("products").Count()

	// Filtered query
	query := database.Table("products")
	if search != "" {
		query = query.Where("name", "LIKE", "%"+search+"%")
	}

	filteredRecords, _ = query.Count()

	query.OrderBy(orderColumn, orderDir).
		Offset(start).
		Limit(length).
		Get(&products)

	response := DataTablesResponse{
		Draw:            draw,
		RecordsTotal:    totalRecords,
		RecordsFiltered: filteredRecords,
		Data:            products,
	}

	p.Ctx.JSON(http.StatusOK, response)
}

// Add displays the add product form
func (p *Product) Add() {
	if !libs.IsLoggedIn(p.Ctx) {
		p.Ctx.Redirect(http.StatusSeeOther, "/auth/login")
		return
	}

	data := core.Map{
		"Title":  "Tambah Product",
		"Values": ProductForm{},
		"Errors": map[string]string{},
	}

	p.Ctx.View("admin/inc/header", data)
	p.Ctx.View("admin/product/add", data)
	p.Ctx.View("admin/inc/footer", data)
}

// Store saves a new product
func (p *Product) Store() {
	if !libs.IsLoggedIn(p.Ctx) {
		p.Ctx.Redirect(http.StatusSeeOther, "/auth/login")
		return
	}

	name := p.Ctx.FormValue("name")
	priceStr := p.Ctx.FormValue("price")
	stockStr := p.Ctx.FormValue("stock")

	price, _ := strconv.ParseFloat(priceStr, 64)
	stock, _ := strconv.Atoi(stockStr)

	errors := make(map[string]string)
	if name == "" || len(name) < 3 {
		errors["Name"] = "Nama product wajib diisi (min 3 karakter)"
	}
	if price <= 0 {
		errors["Price"] = "Harga harus lebih dari 0"
	}

	if len(errors) > 0 {
		data := core.Map{
			"Title":  "Tambah Product",
			"Values": ProductForm{Name: name, Price: price, Stock: stock},
			"Errors": errors,
		}
		p.Ctx.View("admin/inc/header", data)
		p.Ctx.View("admin/product/add", data)
		p.Ctx.View("admin/inc/footer", data)
		return
	}

	now := time.Now()
	database.Table("products").Insert(map[string]any{
		"name":       name,
		"price":      price,
		"stock":      stock,
		"created_at": now,
		"updated_at": now,
	})

	libs.SetFlash(p.Ctx, "success", "Product berhasil ditambahkan")
	p.Ctx.Redirect(http.StatusSeeOther, "/admin/product/index")
}

// Edit displays the edit product form
func (p *Product) Edit() {
	if !libs.IsLoggedIn(p.Ctx) {
		p.Ctx.Redirect(http.StatusSeeOther, "/auth/login")
		return
	}

	id := p.Ctx.Param("id")
	var product models.Product
	err := database.Table("products").Where("id", id).First(&product)
	if err != nil {
		libs.SetFlash(p.Ctx, "error", "Product tidak ditemukan")
		p.Ctx.Redirect(http.StatusSeeOther, "/admin/product/index")
		return
	}

	data := core.Map{
		"Title":   "Edit Product",
		"Product": product,
		"Values": ProductForm{
			Name:  product.Name,
			Price: product.Price,
			Stock: product.Stock,
		},
		"Errors": map[string]string{},
	}

	println(data)

	p.Ctx.View("admin/inc/header", data)
	p.Ctx.View("admin/product/edit", data)
	p.Ctx.View("admin/inc/footer", data)
}

// Update saves product changes
func (p *Product) Update() {
	if !libs.IsLoggedIn(p.Ctx) {
		p.Ctx.Redirect(http.StatusSeeOther, "/auth/login")
		return
	}

	id := p.Ctx.Param("id")
	var product models.Product
	err := database.Table("products").Where("id", id).First(&product)
	if err != nil {
		libs.SetFlash(p.Ctx, "error", "Product tidak ditemukan")
		p.Ctx.Redirect(http.StatusSeeOther, "/admin/product/index")
		return
	}

	name := p.Ctx.FormValue("name")
	priceStr := p.Ctx.FormValue("price")
	stockStr := p.Ctx.FormValue("stock")

	price, _ := strconv.ParseFloat(priceStr, 64)
	stock, _ := strconv.Atoi(stockStr)

	errors := make(map[string]string)
	if name == "" || len(name) < 3 {
		errors["Name"] = "Nama product wajib diisi (min 3 karakter)"
	}
	if price <= 0 {
		errors["Price"] = "Harga harus lebih dari 0"
	}

	if len(errors) > 0 {
		data := core.Map{
			"Title":   "Edit Product",
			"Product": product,
			"Values":  ProductForm{Name: name, Price: price, Stock: stock},
			"Errors":  errors,
		}
		p.Ctx.View("admin/inc/header", data)
		p.Ctx.View("admin/product/edit", data)
		p.Ctx.View("admin/inc/footer", data)
		return
	}

	database.Table("products").Where("id", id).Update(map[string]any{
		"name":       name,
		"price":      price,
		"stock":      stock,
		"updated_at": time.Now(),
	})

	libs.SetFlash(p.Ctx, "success", "Product berhasil diupdate")
	p.Ctx.Redirect(http.StatusSeeOther, "/admin/product/index")
}

// Delete removes a product
func (p *Product) Delete() {
	if !libs.IsLoggedIn(p.Ctx) {
		p.Ctx.JSON(http.StatusUnauthorized, core.Map{"error": "Unauthorized"})
		return
	}

	id := p.Ctx.Param("id")
	err := database.Table("products").Where("id", id).Delete()
	if err != nil {
		p.Ctx.JSON(http.StatusInternalServerError, core.Map{"error": "Gagal menghapus product"})
		return
	}

	p.Ctx.JSON(http.StatusOK, core.Map{"message": "Product berhasil dihapus"})
}
