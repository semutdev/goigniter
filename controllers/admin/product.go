package admin

import (
	"goigniter/config"
	"goigniter/libs"
	"goigniter/models"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func init() {
	libs.Register("product", &Product{})
}

type UserForm struct {
	Name  string `form:"name" validate:"required,min=3"`
	Email string `form:"email" validate:"required,email"`
}

type Product struct{}

func (u *Product) Index(c echo.Context) error {
	var product []models.User
	result := config.DB.Order("created_at desc").Find(&product)

	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "Database Error")
	}

	// send data to view
	data := map[string]interface{}{
		"Title":   "User Management",
		"Product": product,
		"Values":  UserForm{},
		"Errors":  map[string]string{},
	}

	return c.Render(http.StatusOK, "index", data)
}

func (u *Product) Add(c echo.Context) error {
	form := new(UserForm)

	// 1. Tangkap Input
	if err := c.Bind(form); err != nil {
		return c.String(http.StatusBadRequest, "Bad Request")
	}

	// 2. Validasi ($this->form_validation->run())
	if err := c.Validate(form); err != nil {
		// JIKA VALIDASI GAGAL:

		// Konversi error jadi map string
		validationErrors := getValidationErrors(err)

		// Siapkan data untuk dikirim balik ke View (Error + Old Value)
		data := map[string]interface{}{
			"Errors": validationErrors,
			"Values": form, // Mengirim balik apa yang diinput user (set_value)
		}

		// Trik HTMX: Kita suruh HTMX ganti targetnya.
		// Awalnya form targetnya ke tabel (#user-table-body),
		// tapi karena error, kita mau update form-nya saja (#form-container atau parent form)
		// Cara paling gampang di kasus ini: return form-nya saja dengan status 422
		// Dan di frontend kita perlu handle target (atau gunakan `hx-target-error` extension).

		// Tapi cara paling simple tanpa extension:
		// Kita ubah target via Header response
		c.Response().Header().Set("HX-Retarget", "#form-container")
		// Note: Pastikan div pembungkus form di index.html punya id="form-container"

		return c.Render(http.StatusOK, "form_add", data)
	}

	// 3. JIKA SUKSES
	firstName := form.Name
	newUser := models.User{
		FirstName: &firstName,
		Email:     form.Email,
		CreatedOn: time.Now().Unix(),
		Active:    true,
	}

	config.DB.Create(&newUser)

	c.Response().Header().Set("HX-Trigger", "reset-form")

	// Ambil data terbaru
	var product []models.User
	config.DB.Order("created_at desc").Find(&product)

	// Render tabelnya (Target asli form adalah #user-table-body)
	return c.Render(http.StatusOK, "user_list", map[string]interface{}{
		"Product": product,
	})
}

func (u *Product) Delete(c echo.Context) error {
	id := c.Param("id")

	// Hapus data berdasarkan ID
	// Unscoped() digunakan agar benar-benar terhapus (hard delete)
	// Kalau mau soft delete (gorm.DeletedAt), hapus .Unscoped() nya
	if err := config.DB.Unscoped().Delete(&models.User{}, id).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Gagal menghapus")
	}

	// Return 200 OK dengan string kosong.
	// HTMX akan menerima ini dan menghapus elemen <tr> di HTML.
	return c.NoContent(http.StatusOK)
}

func (u *Product) Edit(c echo.Context) error {
	id := c.Param("id")
	var user models.User
	config.DB.First(&user, id)

	// Render file "user_edit_row.html"
	return c.Render(http.StatusOK, "user_edit_row", user)
}

func (u *Product) Row(c echo.Context) error {
	id := c.Param("id")
	var user models.User
	config.DB.First(&user, id)

	// Render potongan template "user_row_only" yang ada di user_list.html
	return c.Render(http.StatusOK, "user_row_only", user)
}

func (u *Product) Update(c echo.Context) error {
	id := c.Param("id")

	var user models.User
	// Cari user dulu
	if err := config.DB.First(&user, id).Error; err != nil {
		return c.String(http.StatusNotFound, "User not found")
	}

	// Update field
	firstName := c.FormValue("name")
	user.FirstName = &firstName
	user.Email = c.FormValue("email")

	config.DB.Save(&user)

	// Setelah save, kembalikan tampilan menjadi baris tabel biasa
	return c.Render(http.StatusOK, "user_row_only", user)
}

func (u *Product) Detail(c echo.Context) error {
	return c.String(http.StatusOK, "Detail user")
}

func getValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			// Custom pesan error ala CI3
			switch fieldError.Tag() {
			case "required":
				errors[fieldError.Field()] = "Field ini wajib diisi bro"
			case "email":
				errors[fieldError.Field()] = "Format email salah"
			case "min":
				errors[fieldError.Field()] = "Minimal " + fieldError.Param() + " karakter"
			default:
				errors[fieldError.Field()] = "Error validasi"
			}
		}
	}
	return errors
}
