package admin

import (
	"full-crud/application/libs"
	"full-crud/application/models"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/semutdev/goigniter/system/core"
	"github.com/semutdev/goigniter/system/libraries/database"
	"github.com/semutdev/goigniter/system/libraries/upload"
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
	Image string  `json:"image"`
}

// DataTablesResponse format response for DataTables
type DataTablesResponse struct {
	Draw            int              `json:"draw"`
	RecordsTotal    int64            `json:"recordsTotal"`
	RecordsFiltered int64            `json:"recordsFiltered"`
	Data            []models.Product `json:"data"`
}

// uploadConfig returns default upload configuration for product images
func uploadConfig() upload.Config {
	return upload.Config{
		UploadPath:   "./public/uploads/products",
		AllowedTypes: "jpg|jpeg|png|gif|webp",
		MaxSize:      2048, // 2MB
		FileName:     "timestamp",
		CreateDirs:   true,
	}
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

	// Handle image upload
	var imageFilename string
	file, header, err := p.Ctx.Request.FormFile("image")
	if err == nil && header != nil {
		file.Close() // Close it, upload library will reopen

		uploader := upload.New(uploadConfig())
		result, err := uploader.Do("image", p.Ctx.Request)
		if err != nil {
			if err == upload.ErrInvalidType {
				errors["Image"] = "Tipe file tidak didukung (hanya jpg, png, gif, webp)"
			} else if err == upload.ErrFileTooBig {
				errors["Image"] = "Ukuran file terlalu besar (max 2MB)"
			} else {
				errors["Image"] = "Gagal mengupload gambar"
			}
		} else {
			imageFilename = result.FileName

			// Create thumbnail (optional)
			imgProcessor := upload.NewImageProcessor(upload.ImageConfig{
				Source:              result.FilePath,
				CreateThumbnail:     true,
				ThumbnailPrefix:     "thumb_",
				ThumbnailWidth:      150,
				ThumbnailHeight:     150,
				MaintainAspectRatio: true,
			})
			imgProcessor.Resize()
		}
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
		"image":      imageFilename,
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
			Image: product.Image,
		},
		"Errors": map[string]string{},
	}

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
	removeImage := p.Ctx.FormValue("remove_image") == "1"

	price, _ := strconv.ParseFloat(priceStr, 64)
	stock, _ := strconv.Atoi(stockStr)

	errors := make(map[string]string)
	if name == "" || len(name) < 3 {
		errors["Name"] = "Nama product wajib diisi (min 3 karakter)"
	}
	if price <= 0 {
		errors["Price"] = "Harga harus lebih dari 0"
	}

	// Handle image upload
	imageFilename := product.Image
	file, header, err := p.Ctx.Request.FormFile("image")
	if err == nil && header != nil {
		file.Close()

		uploader := upload.New(uploadConfig())
		result, err := uploader.Do("image", p.Ctx.Request)
		if err != nil {
			if err == upload.ErrInvalidType {
				errors["Image"] = "Tipe file tidak didukung (hanya jpg, png, gif, webp)"
			} else if err == upload.ErrFileTooBig {
				errors["Image"] = "Ukuran file terlalu besar (max 2MB)"
			} else {
				errors["Image"] = "Gagal mengupload gambar"
			}
		} else {
			// Delete old image if exists
			if product.Image != "" {
				deleteProductImage(product.Image)
			}

			imageFilename = result.FileName

			// Create thumbnail
			imgProcessor := upload.NewImageProcessor(upload.ImageConfig{
				Source:              result.FilePath,
				CreateThumbnail:     true,
				ThumbnailPrefix:     "thumb_",
				ThumbnailWidth:      150,
				ThumbnailHeight:     150,
				MaintainAspectRatio: true,
			})
			imgProcessor.Resize()
		}
	} else if removeImage && product.Image != "" {
		// Remove existing image
		deleteProductImage(product.Image)
		imageFilename = ""
	}

	if len(errors) > 0 {
		data := core.Map{
			"Title":   "Edit Product",
			"Product": product,
			"Values":  ProductForm{Name: name, Price: price, Stock: stock, Image: product.Image},
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
		"image":      imageFilename,
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

	// Get product to delete image
	var product models.Product
	database.Table("products").Where("id", id).First(&product)

	// Delete image files
	if product.Image != "" {
		deleteProductImage(product.Image)
	}

	err := database.Table("products").Where("id", id).Delete()
	if err != nil {
		p.Ctx.JSON(http.StatusInternalServerError, core.Map{"error": "Gagal menghapus product"})
		return
	}

	p.Ctx.JSON(http.StatusOK, core.Map{"message": "Product berhasil dihapus"})
}

// deleteProductImage deletes product image and thumbnail
func deleteProductImage(filename string) {
	if filename == "" {
		return
	}

	uploadPath := "./public/uploads/products"
	imagePath := filepath.Join(uploadPath, filename)
	thumbPath := filepath.Join(uploadPath, "thumb_"+filename)

	os.Remove(imagePath)
	os.Remove(thumbPath)
}

// Helper to check if string is empty
func isEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}