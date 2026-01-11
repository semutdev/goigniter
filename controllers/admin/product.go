package admin

import (
	"goigniter/config"
	"goigniter/libs"
	"goigniter/models"
	"math"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func init() {
	libs.Register("admin/product", &Product{})
}

type Product struct{}

type ProductForm struct {
	Name  string  `form:"name" validate:"required,min=3"`
	Price float64 `form:"price" validate:"required,gt=0"`
	Stock int     `form:"stock" validate:"gte=0"`
}

// DataTablesRequest untuk server-side processing
type DataTablesRequest struct {
	Draw   int    `query:"draw"`
	Start  int    `query:"start"`
	Length int    `query:"length"`
	Search string `query:"search[value]"`
	Order  string `query:"order[0][column]"`
	Dir    string `query:"order[0][dir]"`
}

// DataTablesResponse format response untuk DataTables
type DataTablesResponse struct {
	Draw            int                `json:"draw"`
	RecordsTotal    int64              `json:"recordsTotal"`
	RecordsFiltered int64              `json:"recordsFiltered"`
	Data            []models.Product   `json:"data"`
}

// Index menampilkan halaman list product dengan DataTables
func (p *Product) Index(c echo.Context) error {
	// Cek auth
	if !libs.IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/auth/login")
	}

	data := map[string]interface{}{
		"Title":   "Product Management",
		"Success": libs.GetFlash(c, "success"),
		"Error":   libs.GetFlash(c, "error"),
	}

	return c.Render(http.StatusOK, "admin/product/index", data)
}

// Data mengembalikan JSON untuk DataTables server-side
func (p *Product) Data(c echo.Context) error {
	// Cek auth
	if !libs.IsLoggedIn(c) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	// Parse parameters
	draw, _ := strconv.Atoi(c.QueryParam("draw"))
	start, _ := strconv.Atoi(c.QueryParam("start"))
	length, _ := strconv.Atoi(c.QueryParam("length"))
	search := c.QueryParam("search[value]")
	orderCol := c.QueryParam("order[0][column]")
	orderDir := c.QueryParam("order[0][dir]")

	// Default values
	if length <= 0 {
		length = 10
	}
	if orderDir == "" {
		orderDir = "asc"
	}

	// Column mapping
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

	// Query builder
	var products []models.Product
	var totalRecords int64
	var filteredRecords int64

	// Count total records
	config.DB.Model(&models.Product{}).Count(&totalRecords)

	// Build query with search
	query := config.DB.Model(&models.Product{})
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	// Count filtered records
	query.Count(&filteredRecords)

	// Get data with pagination and ordering
	query.Order(orderColumn + " " + orderDir).
		Offset(start).
		Limit(length).
		Find(&products)

	// Response
	response := DataTablesResponse{
		Draw:            draw,
		RecordsTotal:    totalRecords,
		RecordsFiltered: filteredRecords,
		Data:            products,
	}

	return c.JSON(http.StatusOK, response)
}

// Add menampilkan form tambah product
func (p *Product) Add(c echo.Context) error {
	if !libs.IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/auth/login")
	}

	data := map[string]interface{}{
		"Title":  "Tambah Product",
		"Values": ProductForm{},
		"Errors": map[string]string{},
	}

	return c.Render(http.StatusOK, "admin/product/add", data)
}

// Store menyimpan product baru
func (p *Product) Store(c echo.Context) error {
	if !libs.IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/auth/login")
	}

	form := new(ProductForm)
	if err := c.Bind(form); err != nil {
		libs.SetFlash(c, "error", "Data tidak valid")
		return c.Redirect(http.StatusSeeOther, "/admin/product/add")
	}

	if err := c.Validate(form); err != nil {
		data := map[string]interface{}{
			"Title":  "Tambah Product",
			"Values": form,
			"Errors": getProductValidationErrors(err),
		}
		return c.Render(http.StatusOK, "admin/product/add", data)
	}

	// Simpan ke database
	product := models.Product{
		Name:  form.Name,
		Price: form.Price,
		Stock: form.Stock,
	}
	config.DB.Create(&product)

	libs.SetFlash(c, "success", "Product berhasil ditambahkan")
	return c.Redirect(http.StatusSeeOther, "/admin/product")
}

// Edit menampilkan form edit product
func (p *Product) Edit(c echo.Context) error {
	if !libs.IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/auth/login")
	}

	id := c.Param("id")
	var product models.Product
	if err := config.DB.First(&product, id).Error; err != nil {
		libs.SetFlash(c, "error", "Product tidak ditemukan")
		return c.Redirect(http.StatusSeeOther, "/admin/product")
	}

	data := map[string]interface{}{
		"Title":   "Edit Product",
		"Product": product,
		"Values": ProductForm{
			Name:  product.Name,
			Price: product.Price,
			Stock: product.Stock,
		},
		"Errors": map[string]string{},
	}

	return c.Render(http.StatusOK, "admin/product/edit", data)
}

// Update menyimpan perubahan product
func (p *Product) Update(c echo.Context) error {
	if !libs.IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/auth/login")
	}

	id := c.Param("id")
	var product models.Product
	if err := config.DB.First(&product, id).Error; err != nil {
		libs.SetFlash(c, "error", "Product tidak ditemukan")
		return c.Redirect(http.StatusSeeOther, "/admin/product")
	}

	form := new(ProductForm)
	if err := c.Bind(form); err != nil {
		libs.SetFlash(c, "error", "Data tidak valid")
		return c.Redirect(http.StatusSeeOther, "/admin/product/edit/"+id)
	}

	if err := c.Validate(form); err != nil {
		data := map[string]interface{}{
			"Title":   "Edit Product",
			"Product": product,
			"Values":  form,
			"Errors":  getProductValidationErrors(err),
		}
		return c.Render(http.StatusOK, "admin/product/edit", data)
	}

	// Update database
	product.Name = form.Name
	product.Price = form.Price
	product.Stock = form.Stock
	config.DB.Save(&product)

	libs.SetFlash(c, "success", "Product berhasil diupdate")
	return c.Redirect(http.StatusSeeOther, "/admin/product")
}

// Delete menghapus product
func (p *Product) Delete(c echo.Context) error {
	if !libs.IsLoggedIn(c) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	id := c.Param("id")
	if err := config.DB.Delete(&models.Product{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal menghapus product"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Product berhasil dihapus"})
}

// Helper untuk format harga
func formatPrice(price float64) string {
	return strconv.FormatFloat(math.Round(price*100)/100, 'f', 0, 64)
}

func getProductValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(interface{ Field() string }); ok {
		_ = validationErrors
	}

	// Simple error extraction
	if err != nil {
		errStr := err.Error()
		if contains(errStr, "Name") {
			errors["Name"] = "Nama product wajib diisi (min 3 karakter)"
		}
		if contains(errStr, "Price") {
			errors["Price"] = "Harga harus lebih dari 0"
		}
		if contains(errStr, "Stock") {
			errors["Stock"] = "Stock tidak boleh negatif"
		}
	}
	return errors
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsRune(s, substr))
}

func containsRune(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
