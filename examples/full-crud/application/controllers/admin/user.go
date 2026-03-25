package admin

import (
	"full-crud/application/libs"
	"full-crud/application/models"
	"net/http"
	"strconv"
	"time"

	"github.com/semutdev/goigniter/system/core"
	"github.com/semutdev/goigniter/system/libraries/database"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	core.Register(&User{}, "admin")
}

type User struct {
	core.Controller
}

// Routes defines custom routes for User controller
func (u *User) Routes() map[string]string {
	return map[string]string{
		"Edit":     "edit/:id",
		"Update":   "update/:id",
		"Delete":   "delete/:id",
		"Activate": "activate/:id",
	}
}

type UserForm struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Company   string `json:"company"`
	Phone     string `json:"phone"`
	Active    bool   `json:"active"`
}

// isHtmx checks if the request is from HTMX
func isHtmx(c *core.Context) bool {
	return c.Header("HX-Request") == "true"
}

// Index displays the user list page
func (u *User) Index() {
	if !libs.RequireGroup(u.Ctx, "admin") {
		return
	}

	var users []models.User
	database.Table("users").OrderBy("id", "desc").Get(&users)

	// Load groups for each user
	for i := range users {
		loadUserGroupsAdmin(&users[i])
	}

	data := core.Map{
		"Title":   "User Management",
		"Users":   users,
		"Success": libs.GetFlash(u.Ctx, "success"),
		"Error":   libs.GetFlash(u.Ctx, "error"),
	}

	u.Ctx.View("admin/inc/header", data)
	u.Ctx.View("admin/user/index", data)
	u.Ctx.View("admin/inc/footer", data)
}

// Add displays the add user form
func (u *User) Add() {
	if !libs.RequireGroup(u.Ctx, "admin") {
		return
	}

	data := core.Map{
		"Title":  "Tambah User",
		"Values": UserForm{},
		"Errors": map[string]string{},
	}

	// If HTMX request, return just the form partial
	if isHtmx(u.Ctx) {
		u.Ctx.View("admin/user/_form", data)
		return
	}

	u.Ctx.View("admin/inc/header", data)
	u.Ctx.View("admin/user/add", data)
	u.Ctx.View("admin/inc/footer", data)
}

// Store saves a new user
func (u *User) Store() {
	if !libs.RequireGroup(u.Ctx, "admin") {
		return
	}

	email := u.Ctx.FormValue("email")
	password := u.Ctx.FormValue("password")
	firstName := u.Ctx.FormValue("first_name")
	lastName := u.Ctx.FormValue("last_name")
	company := u.Ctx.FormValue("company")
	phone := u.Ctx.FormValue("phone")
	active := u.Ctx.FormValue("active") == "1"

	errors := make(map[string]string)
	if email == "" {
		errors["Email"] = "Email wajib diisi"
	}
	if password == "" || len(password) < 6 {
		errors["Password"] = "Password wajib diisi (min 6 karakter)"
	}

	// Check if email already exists
	var existingUser models.User
	err := database.Table("users").Where("email", email).First(&existingUser)
	if err == nil {
		errors["Email"] = "Email sudah terdaftar"
	}

	if len(errors) > 0 {
		data := core.Map{
			"Title":  "Tambah User",
			"Values": UserForm{
				Email:     email,
				Password:  password,
				FirstName: firstName,
				LastName:  lastName,
				Company:   company,
				Phone:     phone,
				Active:    active,
			},
			"Errors": errors,
		}

		// If HTMX request, return form with errors
		if isHtmx(u.Ctx) {
			u.Ctx.View("admin/user/_form", data)
			return
		}

		u.Ctx.View("admin/inc/header", data)
		u.Ctx.View("admin/user/add", data)
		u.Ctx.View("admin/inc/footer", data)
		return
	}

	// Hash password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	now := time.Now().Unix()
	userID, err := database.Table("users").InsertGetId(map[string]any{
		"email":      email,
		"password":   string(hashedPassword),
		"first_name": firstName,
		"last_name":  lastName,
		"company":    company,
		"phone":      phone,
		"ip_address": u.Ctx.Request.RemoteAddr,
		"created_on": now,
		"active":     active,
	})

	if err != nil {
		// If HTMX request, return form with error
		if isHtmx(u.Ctx) {
			data := core.Map{
				"Title":  "Tambah User",
				"Values": UserForm{Email: email, FirstName: firstName, LastName: lastName, Company: company, Phone: phone, Active: active},
				"Errors": map[string]string{"Email": "Gagal menyimpan user"},
			}
			u.Ctx.View("admin/user/_form", data)
			return
		}

		libs.SetFlash(u.Ctx, "error", "Gagal menyimpan user")
		u.Ctx.Redirect(http.StatusSeeOther, "/admin/user/index")
		return
	}

	// Get members group ID
	var membersGroupID int64
	database.Table("groups").Where("name", "members").Select("id").First(&membersGroupID)

	// Assign to members group
	database.Table("users_groups").Insert(map[string]any{
		"user_id":  userID,
		"group_id": membersGroupID,
	})

	// If HTMX request, return redirect header
	if isHtmx(u.Ctx) {
		u.Ctx.SetHeader("HX-Redirect", "/admin/user/index")
		u.Ctx.String(http.StatusOK, "")
		return
	}

	libs.SetFlash(u.Ctx, "success", "User berhasil ditambahkan")
	u.Ctx.Redirect(http.StatusSeeOther, "/admin/user/index")
}

// Edit displays the edit user form
func (u *User) Edit() {
	if !libs.RequireGroup(u.Ctx, "admin") {
		return
	}

	id := u.Ctx.Param("id")
	var user models.User
	err := database.Table("users").Where("id", id).First(&user)
	if err != nil {
		libs.SetFlash(u.Ctx, "error", "User tidak ditemukan")
		u.Ctx.Redirect(http.StatusSeeOther, "/admin/user/index")
		return
	}

	// Load groups
	loadUserGroupsAdmin(&user)

	data := core.Map{
		"Title": "Edit User",
		"User":  user,
		"Values": UserForm{
			Email:     user.Email,
			FirstName: getString(user.FirstName),
			LastName:  getString(user.LastName),
			Company:   getString(user.Company),
			Phone:     getString(user.Phone),
			Active:    user.Active,
		},
		"Errors": map[string]string{},
	}

	u.Ctx.View("admin/inc/header", data)
	u.Ctx.View("admin/user/edit", data)
	u.Ctx.View("admin/inc/footer", data)
}

// Update saves user changes
func (u *User) Update() {
	if !libs.RequireGroup(u.Ctx, "admin") {
		return
	}

	id := u.Ctx.Param("id")
	var user models.User
	err := database.Table("users").Where("id", id).First(&user)
	if err != nil {
		libs.SetFlash(u.Ctx, "error", "User tidak ditemukan")
		u.Ctx.Redirect(http.StatusSeeOther, "/admin/user/index")
		return
	}

	email := u.Ctx.FormValue("email")
	password := u.Ctx.FormValue("password")
	firstName := u.Ctx.FormValue("first_name")
	lastName := u.Ctx.FormValue("last_name")
	company := u.Ctx.FormValue("company")
	phone := u.Ctx.FormValue("phone")
	active := u.Ctx.FormValue("active") == "1"

	errors := make(map[string]string)
	if email == "" {
		errors["Email"] = "Email wajib diisi"
	}

	// Check if email already exists (other user)
	var existingUser models.User
	err = database.Table("users").Where("email", email).Where("id !=", id).First(&existingUser)
	if err == nil {
		errors["Email"] = "Email sudah digunakan user lain"
	}

	if password != "" && len(password) < 6 {
		errors["Password"] = "Password minimal 6 karakter"
	}

	if len(errors) > 0 {
		data := core.Map{
			"Title": "Edit User",
			"User":  user,
			"Values": UserForm{
				Email:     email,
				Password:  password,
				FirstName: firstName,
				LastName:  lastName,
				Company:   company,
				Phone:     phone,
				Active:    active,
			},
			"Errors": errors,
		}
		u.Ctx.View("admin/inc/header", data)
		u.Ctx.View("admin/user/edit", data)
		u.Ctx.View("admin/inc/footer", data)
		return
	}

	updateData := map[string]any{
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
		"company":    company,
		"phone":      phone,
		"active":     active,
	}

	// Update password only if provided
	if password != "" {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		updateData["password"] = string(hashedPassword)
	}

	database.Table("users").Where("id", id).Update(updateData)

	libs.SetFlash(u.Ctx, "success", "User berhasil diupdate")
	u.Ctx.Redirect(http.StatusSeeOther, "/admin/user/index")
}

// Delete removes a user
func (u *User) Delete() {
	if !libs.RequireGroup(u.Ctx, "admin") {
		u.Ctx.JSON(http.StatusUnauthorized, core.Map{"error": "Unauthorized"})
		return
	}

	id := u.Ctx.Param("id")

	// Prevent deleting own account
	currentUser := libs.GetUser(u.Ctx)
	if currentUser != nil && strconv.FormatInt(currentUser.ID, 10) == id {
		u.Ctx.JSON(http.StatusBadRequest, core.Map{"error": "Tidak dapat menghapus akun sendiri"})
		return
	}

	// Delete from users_groups first
	database.Table("users_groups").Where("user_id", id).Delete()

	// Delete user
	err := database.Table("users").Where("id", id).Delete()
	if err != nil {
		u.Ctx.JSON(http.StatusInternalServerError, core.Map{"error": "Gagal menghapus user"})
		return
	}

	u.Ctx.JSON(http.StatusOK, core.Map{"message": "User berhasil dihapus"})
}

// Activate toggles user active status
func (u *User) Activate() {
	if !libs.RequireGroup(u.Ctx, "admin") {
		u.Ctx.JSON(http.StatusUnauthorized, core.Map{"error": "Unauthorized"})
		return
	}

	id := u.Ctx.Param("id")

	var user models.User
	err := database.Table("users").Where("id", id).First(&user)
	if err != nil {
		u.Ctx.JSON(http.StatusNotFound, core.Map{"error": "User tidak ditemukan"})
		return
	}

	// Toggle active status
	newStatus := !user.Active
	database.Table("users").Where("id", id).Update(map[string]any{
		"active": newStatus,
	})

	status := "dinonaktifkan"
	if newStatus {
		status = "diaktifkan"
	}

	u.Ctx.JSON(http.StatusOK, core.Map{
		"message": "User berhasil " + status,
		"active":  newStatus,
	})
}

// Helper functions
func loadUserGroupsAdmin(user *models.User) {
	var groups []models.Group
	database.Query(`
		SELECT g.id, g.name, g.description
		FROM groups g
		INNER JOIN users_groups ug ON g.id = ug.group_id
		WHERE ug.user_id = ?
	`, user.ID).Get(&groups)
	user.Groups = groups
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}