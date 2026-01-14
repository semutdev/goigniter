package admin

import (
	"full-crud/application/libs"
	"full-crud/application/models"
	"full-crud/config"
	"net/http"
	"strconv"

	"github.com/semutdev/goigniter/system/core"
)

func init() {
	core.Register(&ProductController{}, "admin")
}

type ProductController struct {
	core.Controller
}

type ProductForm struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// DataTablesResponse format response untuk DataTables
type DataTablesResponse struct {
	Draw            int              `json:"draw"`
	RecordsTotal    int64            `json:"recordsTotal"`
	RecordsFiltered int64            `json:"recordsFiltered"`
	Data            []models.Product `json:"data"`
}

// Index menampilkan halaman list product
func (p *ProductController) Index() {
	if !libs.IsLoggedIn(p.Ctx) {
		p.Ctx.Redirect(http.StatusSeeOther, "/auth/login")
		return
	}

	p.Ctx.View("admin/product/index", core.Map{
		"Title":   "Product Management",
		"Success": libs.GetFlash(p.Ctx, "success"),
		"Error":   libs.GetFlash(p.Ctx, "error"),
	})
}

// Data mengembalikan JSON untuk DataTables
func (p *ProductController) Data() {
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

	config.DB.Model(&models.Product{}).Count(&totalRecords)

	query := config.DB.Model(&models.Product{})
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	query.Count(&filteredRecords)

	query.Order(orderColumn + " " + orderDir).
		Offset(start).
		Limit(length).
		Find(&products)

	response := DataTablesResponse{
		Draw:            draw,
		RecordsTotal:    totalRecords,
		RecordsFiltered: filteredRecords,
		Data:            products,
	}

	p.Ctx.JSON(http.StatusOK, response)
}

// Add menampilkan form tambah product
func (p *ProductController) Add() {
	if !libs.IsLoggedIn(p.Ctx) {
		p.Ctx.Redirect(http.StatusSeeOther, "/auth/login")
		return
	}

	p.Ctx.View("admin/product/add", core.Map{
		"Title":  "Tambah Product",
		"Values": ProductForm{},
		"Errors": map[string]string{},
	})
}

// Store menyimpan product baru
func (p *ProductController) Store() {
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
		p.Ctx.View("admin/product/add", core.Map{
			"Title":  "Tambah Product",
			"Values": ProductForm{Name: name, Price: price, Stock: stock},
			"Errors": errors,
		})
		return
	}

	product := models.Product{
		Name:  name,
		Price: price,
		Stock: stock,
	}
	config.DB.Create(&product)

	libs.SetFlash(p.Ctx, "success", "Product berhasil ditambahkan")
	p.Ctx.Redirect(http.StatusSeeOther, "/admin/productcontroller")
}

// Edit menampilkan form edit product
func (p *ProductController) Edit() {
	if !libs.IsLoggedIn(p.Ctx) {
		p.Ctx.Redirect(http.StatusSeeOther, "/auth/login")
		return
	}

	id := p.Ctx.Param("id")
	var product models.Product
	if err := config.DB.First(&product, id).Error; err != nil {
		libs.SetFlash(p.Ctx, "error", "Product tidak ditemukan")
		p.Ctx.Redirect(http.StatusSeeOther, "/admin/productcontroller")
		return
	}

	p.Ctx.View("admin/product/edit", core.Map{
		"Title":   "Edit Product",
		"Product": product,
		"Values": ProductForm{
			Name:  product.Name,
			Price: product.Price,
			Stock: product.Stock,
		},
		"Errors": map[string]string{},
	})
}

// Update menyimpan perubahan product
func (p *ProductController) Update() {
	if !libs.IsLoggedIn(p.Ctx) {
		p.Ctx.Redirect(http.StatusSeeOther, "/auth/login")
		return
	}

	id := p.Ctx.Param("id")
	var product models.Product
	if err := config.DB.First(&product, id).Error; err != nil {
		libs.SetFlash(p.Ctx, "error", "Product tidak ditemukan")
		p.Ctx.Redirect(http.StatusSeeOther, "/admin/productcontroller")
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
		p.Ctx.View("admin/product/edit", core.Map{
			"Title":   "Edit Product",
			"Product": product,
			"Values":  ProductForm{Name: name, Price: price, Stock: stock},
			"Errors":  errors,
		})
		return
	}

	product.Name = name
	product.Price = price
	product.Stock = stock
	config.DB.Save(&product)

	libs.SetFlash(p.Ctx, "success", "Product berhasil diupdate")
	p.Ctx.Redirect(http.StatusSeeOther, "/admin/productcontroller")
}

// Delete menghapus product
func (p *ProductController) Delete() {
	if !libs.IsLoggedIn(p.Ctx) {
		p.Ctx.JSON(http.StatusUnauthorized, core.Map{"error": "Unauthorized"})
		return
	}

	id := p.Ctx.Param("id")
	if err := config.DB.Delete(&models.Product{}, id).Error; err != nil {
		p.Ctx.JSON(http.StatusInternalServerError, core.Map{"error": "Gagal menghapus product"})
		return
	}

	p.Ctx.JSON(http.StatusOK, core.Map{"message": "Product berhasil dihapus"})
}
