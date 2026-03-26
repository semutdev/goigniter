package controllers

import (
	"myapp/application/models"
	"net/http"
	"strconv"

	"github.com/semutdev/goigniter/system/core"
)

func init() {
	core.Register(&User{})
}

// User controller (thin controller)
type User struct {
	core.Controller
}

// Routes defines custom routes with parameters
func (u *User) Routes() map[string]string {
	return map[string]string{
		"Show":   "show/:id",
		"Update": "update/:id",
		"Edit":   "edit/:id",
		"Delete": "delete/:id",
	}
}

// Index displays list of users
func (u *User) Index() {
	userModel := models.NewUserModel()
	users, err := userModel.All()
	if err != nil {
		u.Ctx.String(http.StatusInternalServerError, "Failed to load users")
		return
	}

	u.Ctx.View("user/index", core.Map{
		"Title": "Users",
		"Users": users,
	})
}

// Add displays add user form
func (u *User) Add() {
	u.Ctx.View("user/add", core.Map{
		"Title":  "Add User",
		"Errors": map[string]string{},
	})
}

// Store creates a new user
func (u *User) Store() {
	name := u.Ctx.FormValue("name")
	email := u.Ctx.FormValue("email")

	// Validation
	errors := make(map[string]string)
	if name == "" {
		errors["Name"] = "Name is required"
	}
	if email == "" {
		errors["Email"] = "Email is required"
	}

	userModel := models.NewUserModel()
	if exists, _ := userModel.Exists(email); exists {
		errors["Email"] = "Email already exists"
	}

	if len(errors) > 0 {
		u.Ctx.View("user/add", core.Map{
			"Title":  "Add User",
			"Errors": errors,
			"Name":   name,
			"Email":  email,
		})
		return
	}

	// Create user (business logic in model)
	_, err := userModel.Create(name, email)
	if err != nil {
		u.Ctx.String(http.StatusInternalServerError, "Failed to create user")
		return
	}

	u.Ctx.Redirect(http.StatusSeeOther, "/user/index")
}

// Show displays a single user
func (u *User) Show() {
	id, _ := strconv.ParseInt(u.Ctx.Param("id"), 10, 64)

	userModel := models.NewUserModel()
	user, err := userModel.Find(id)
	if err != nil {
		u.Ctx.String(http.StatusNotFound, "User not found")
		return
	}

	u.Ctx.View("user/show", core.Map{
		"Title": "User Detail",
		"User":  user,
	})
}

// Edit displays edit user form
func (u *User) Edit() {
	id, _ := strconv.ParseInt(u.Ctx.Param("id"), 10, 64)

	userModel := models.NewUserModel()
	user, err := userModel.Find(id)
	if err != nil {
		u.Ctx.String(http.StatusNotFound, "User not found")
		return
	}

	u.Ctx.View("user/edit", core.Map{
		"Title":  "Edit User",
		"User":   user,
		"Errors": map[string]string{},
	})
}

// Update saves user changes
func (u *User) Update() {
	id, _ := strconv.ParseInt(u.Ctx.Param("id"), 10, 64)
	name := u.Ctx.FormValue("name")
	email := u.Ctx.FormValue("email")

	// Validation
	errors := make(map[string]string)
	if name == "" {
		errors["Name"] = "Name is required"
	}
	if email == "" {
		errors["Email"] = "Email is required"
	}

	userModel := models.NewUserModel()

	// Check if user exists
	user, err := userModel.Find(id)
	if err != nil {
		u.Ctx.String(http.StatusNotFound, "User not found")
		return
	}

	// Check email uniqueness (exclude current user)
	if exists, _ := userModel.Exists(email, id); exists {
		errors["Email"] = "Email already exists"
	}

	if len(errors) > 0 {
		u.Ctx.View("user/edit", core.Map{
			"Title":  "Edit User",
			"User":   user,
			"Errors": errors,
		})
		return
	}

	// Update user (business logic in model)
	err = userModel.Update(id, name, email)
	if err != nil {
		u.Ctx.String(http.StatusInternalServerError, "Failed to update user")
		return
	}

	u.Ctx.Redirect(http.StatusSeeOther, "/user/index")
}

// Delete removes a user
func (u *User) Delete() {
	id, _ := strconv.ParseInt(u.Ctx.Param("id"), 10, 64)

	userModel := models.NewUserModel()
	err := userModel.Delete(id)
	if err != nil {
		u.Ctx.JSON(http.StatusInternalServerError, core.Map{"error": "Failed to delete user"})
		return
	}

	u.Ctx.JSON(http.StatusOK, core.Map{"message": "User deleted"})
}
