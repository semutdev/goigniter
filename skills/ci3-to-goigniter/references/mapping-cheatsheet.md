# CodeIgniter 3 to GoIgniter API Mapping Cheatsheet

This cheatsheet provides an exhaustive side-by-side API dictionary and migration reference for converting PHP CodeIgniter 3 (CI3) applications to the GoIgniter Go framework (`github.com/semutdev/goigniter`).

---

## 1. Core Architecture & Philosophy

| Aspect | CodeIgniter 3 (PHP) | GoIgniter (Go) |
|---|---|---|
| **Language Paradigm** | Dynamic, interpreted PHP, class inheritance | Static, compiled Go, struct composition & interfaces |
| **HTTP Engine** | Apache / Nginx + PHP-FPM (`index.php`) | Native standalone HTTP server (`net/http` stdlib) |
| **Core Instance** | Global super-object `$CI =& get_instance()` or `$this->...` | Explicit `*core.Context` passed to handlers, `*core.Application` |
| **Controller Base** | `class Controller extends CI_Controller` | `type Controller struct { core.Controller }` |
| **Package Imports** | `$this->load->library()`, `$this->load->helper()` | Explicit Go imports (e.g. `goigniter/system/core`, `goigniter/system/libraries/...`) |
| **Data Flow** | Associative arrays (`$data['title'] = '...'`) | Structs or `core.Map{"Title": "..."}` (`map[string]any`) |

---

## 2. Routing & HTTP Methods

In CI3, routes are defined in `application/config/routes.php`. In GoIgniter, routes are registered on `*core.Application` or handled via `app.AutoRoute()`.

### 2.1 Route Definitions

```php
// CI3: application/config/routes.php
$route['default_controller'] = 'welcome';
$route['404_override'] = 'errors/page_missing';

$route['users'] = 'user/index';
$route['users/create'] = 'user/create';
$route['users/(:num)'] = 'user/detail/$1';
$route['products/(:any)'] = 'product/view/$1';
$route['blog/(:num)/(:any)'] = 'blog/post/$1/$2';
$route['api/v1/status'] = 'api/status';
```

```go
// GoIgniter: main.go or routes configuration
package main

import (
    "net/http"
    "github.com/semutdev/goigniter/system/core"
)

func RegisterRoutes(app *core.Application) {
    // Explicit route handlers
    app.GET("/", welcomeHandler)
    app.GET("/users", userIndexHandler)
    app.POST("/users", userStoreHandler)
    app.GET("/users/create", userCreateHandler)

    // Path parameters (:id, :slug)
    app.GET("/users/:id", userDetailHandler)
    app.GET("/products/:slug", productViewHandler)
    app.GET("/blog/:id/:slug", blogPostHandler)

    // Wildcards (*filepath, *path)
    app.GET("/files/*filepath", fileServeHandler)

    // Static file serving (e.g., assets)
    app.Static("/assets", "./public/assets")
}
```

### 2.2 HTTP Methods

GoIgniter supports all standard HTTP methods directly:

| HTTP Verb | CI3 (manual check inside method) | GoIgniter API |
|---|---|---|
| **GET** | `$this->input->method() == 'get'` | `app.GET(pattern, handler)` |
| **POST** | `$this->input->method() == 'post'` | `app.POST(pattern, handler)` |
| **PUT** | manual check via PHP input stream | `app.PUT(pattern, handler)` |
| **DELETE** | manual check via `$_SERVER['REQUEST_METHOD']` | `app.DELETE(pattern, handler)` |
| **PATCH** | manual check | `app.PATCH(pattern, handler)` |
| **OPTIONS** | manual check | `app.OPTIONS(pattern, handler)` |
| **HEAD** | manual check | `app.HEAD(pattern, handler)` |

### 2.3 Route Groups

```php
// CI3: No native route groups (requires custom prefix routing or subdirectories)
$route['admin/users'] = 'admin/users/index';
$route['admin/users/create'] = 'admin/users/create';
$route['admin/reports'] = 'admin/reports/index';
```

```go
// GoIgniter: Native nested route groups with middleware
admin := app.Group("/admin", adminAuthMiddleware)
{
    admin.GET("/dashboard", adminDashboardHandler)
    
    users := admin.Group("/users")
    {
        users.GET("", adminUsersIndex)
        users.GET("/create", adminUsersCreate)
        users.POST("", adminUsersStore)
    }

    admin.GET("/reports", adminReportsIndex)
}
```

### 2.4 Auto-Routing

GoIgniter features reflection-based auto-routing mirroring CodeIgniter's convention:

```php
// CI3: Automatic URL resolution
// URL: /users/detail/42 -> Users controller -> detail($id = 42)
class Users extends CI_Controller {
    public function index() { ... }
    public function detail($id) { ... }
}
```

```go
// GoIgniter: Register controller struct and activate AutoRoute
package main

import "github.com/semutdev/goigniter/system/core"

type UserController struct {
    core.Controller
}

func (u *UserController) Index() {
    u.Ctx.JSON(200, core.Map{"status": "ok"})
}

func (u *UserController) Detail() {
    id := u.Ctx.Param("id")
    u.Ctx.JSON(200, core.Map{"id": id})
}

// Custom route parameter patterns for AutoRoute
func (u *UserController) Routes() map[string]string {
    return map[string]string{
        "Detail": "detail/:id",
    }
}

// Restrict HTTP methods per action
func (u *UserController) AllowedMethods() map[string][]string {
    return map[string][]string{
        "Index":  {"GET"},
        "Detail": {"GET"},
    }
}

func main() {
    app := core.New()
    core.Register(&UserController{})
    // Or with prefix: core.Register(&AdminController{}, "admin")
    app.AutoRoute()
    app.Run(":8080")
}
```

---

## 3. Controllers & Action Lifecycle

### 3.1 Controller Struct & Actions

```php
// CI3: application/controllers/User.php
<?php
defined('BASEPATH') OR exit('No direct script access allowed');

class User extends CI_Controller {
    public function __construct() {
        parent::__construct();
        $this->load->model('User_model');
        $this->load->helper('url');
    }

    public function index() {
        $data['users'] = $this->User_model->get_all();
        $this->load->view('users/index', $data);
    }
}
```

```go
// GoIgniter: controllers/user.go
package controllers

import (
    "net/http"
    "github.com/semutdev/goigniter/system/core"
    "myapp/models"
)

type UserController struct {
    core.Controller
    userModel *models.UserModel
}

func NewUserController() *UserController {
    return &UserController{
        userModel: models.NewUserModel(),
    }
}

// Action method using embedded core.Controller (AutoRoute style)
func (u *UserController) Index() {
    users, err := u.userModel.GetAll()
    if err != nil {
        u.Ctx.String(http.StatusInternalServerError, err.Error())
        return
    }

    u.Ctx.View("users/index", core.Map{
        "Users": users,
    })
}

// Or as a standalone HandlerFunc (Express/Echo style)
func (u *UserController) IndexHandler(c *core.Context) error {
    users, err := u.userModel.GetAll()
    if err != nil {
        return c.String(http.StatusInternalServerError, err.Error())
    }
    return c.View("users/index", core.Map{
        "Users": users,
    })
}
```

### 3.2 Action Middleware Hooks

```go
// GoIgniter controllers can define middleware per controller or per method:
func (u *UserController) Middleware() []core.Middleware {
    return []core.Middleware{
        // Controller-wide middleware
    }
}

func (u *UserController) MiddlewareFor() map[string][]core.Middleware {
    return map[string][]core.Middleware{
        "Secret": {authMiddleware},
    }
}
```

---

## 4. Request & Input Handling

In CI3, user input is accessed via `$this->input`. In GoIgniter, methods on `c *core.Context` provide clean, type-aware access.

### 4.1 Input Mapping Summary

| CI3 Input Method | GoIgniter Context Method | Description / Return Type |
|---|---|---|
| `$this->input->post('username')` | `c.Form("username")` | Returns `string` (or `c.FormValue`) |
| `$this->input->get('page')` | `c.Query("page")` | Returns `string` query param |
| `$this->input->get('page') ?: '1'` | `c.QueryDefault("page", "1")` | Returns `string` with default fallback |
| `(int)$this->input->get('page')` | `c.QueryInt("page")` | Returns `(int, error)` |
| `(int)$this->input->get('page') ?: 1`| `c.QueryIntDefault("page", 1)` | Returns `int` with default fallback |
| `$this->uri->segment(3)` | `c.Param("id")` | Returns `string` route parameter |
| `(int)$this->uri->segment(3)` | `c.ParamInt("id")` | Returns `(int, error)` route parameter |
| `$this->input->server('HTTP_HOST')` | `c.Header("Host")` | Reads incoming request header |
| `$this->input->get_request_header('X')` | `c.Header("X")` | Reads incoming request header |
| `$this->input->ip_address()` | `c.IP()` | Returns IP (handles `X-Forwarded-For`) |
| `$this->input->method()` | `c.Method()` | Returns HTTP method string (e.g. `"POST"`) |
| `$this->uri->uri_string()` | `c.Path()` | Returns request URL path string |
| `$this->input->cookie('session_id')`| `cookie, err := c.Cookie("name")`| Returns `*http.Cookie` |
| `set_cookie('name', 'val', 3600)` | `c.SetCookie(&http.Cookie{...})` | Sets outgoing HTTP cookie |

### 4.2 JSON Body & Binding

```php
// CI3: Read raw JSON body
$stream_clean = $this->security->xss_clean($this->input->raw_input_stream);
$request_body = json_decode($stream_clean, true);
$email = $request_body['email'] ?? '';
```

```go
// GoIgniter: Type-safe JSON body binding
type CreateUserRequest struct {
    Email    string `json:"email"`
    Username string `json:"username"`
    Role     string `json:"role"`
}

func (u *UserController) Store() {
    var req CreateUserRequest

    // Bind with automatic content-type inspection (JSON or form-urlencoded)
    if err := u.Ctx.Bind(&req); err != nil {
        u.Ctx.JSON(http.StatusBadRequest, core.Map{"error": "Invalid request body"})
        return
    }

    // Or bind with a strict byte limit to prevent memory exhaustion attacks:
    // if err := u.Ctx.BindWithLimit(&req, 1024*1024); err != nil { ... }
}
```

### 4.3 Context-Scoped Storage

```php
// CI3: Passing data through hooks/controllers
$this->load->vars(['current_user' => $user]);
```

```go
// GoIgniter: Request-scoped store (useful across middleware)
c.Set("user_id", 42)
c.Set("user_role", "admin")

// Retrieval
userID := c.GetInt("user_id")
role := c.GetString("user_role")
rawVal := c.Get("any_key")
```

---

## 5. Response Handling

In CI3, response rendering is governed by `$this->output` or direct echos. In GoIgniter, `c *core.Context` offers standardized response helpers.

### 5.1 Response Helper Matrix

| CI3 Response Method | GoIgniter Equivalent | Content-Type |
|---|---|---|
| `$this->output->set_content_type('application/json')->set_output(json_encode($data));` | `c.JSON(http.StatusOK, data)` | `application/json; charset=utf-8` |
| `$this->output->set_content_type('text/plain')->set_output($text);` | `c.String(http.StatusOK, "text")` | `text/plain; charset=utf-8` |
| `$this->output->set_output($html);` | `c.HTML(http.StatusOK, htmlString)` | `text/html; charset=utf-8` |
| `$this->load->view('page', $data);` | `c.View("page", dataMap)` | `text/html; charset=utf-8` |
| `redirect('/login');` | `c.Redirect(http.StatusFound, "/login")` | `Location: /login` |
| `redirect('/target', 'location', 301);` | `c.Redirect(http.StatusMovedPermanently, "/target")` | `Location: /target` |
| `force_download('report.pdf', $binary);` | `c.File("./storage/report.pdf")` | Auto-detected MIME |
| `$this->output->set_header('Content-Type: image/png'); echo $img;` | `c.Blob(http.StatusOK, "image/png", bytes)` | Custom binary |
| `$this->output->set_status_header(204);` | `c.NoContent(http.StatusNoContent)` | Empty body |

### 5.2 Setting Custom Headers

```php
// CI3
$this->output->set_header('X-Custom-Header: FooBar');
$this->output->set_status_header(422);
```

```go
// GoIgniter
c.SetHeader("X-Custom-Header", "FooBar")
c.JSON(http.StatusUnprocessableEntity, core.Map{"error": "Validation failed"})
```

---

## 6. Session Management & Flashdata

CI3 uses either file or database session drivers. GoIgniter provides cookie-based sessions secured with **HMAC-SHA256 signature verification** and optional **AES-256-GCM encryption**.

### 6.1 Session Setup & Initialization

```php
// CI3: application/config/config.php
$config['sess_driver'] = 'files';
$config['sess_cookie_name'] = 'ci_session';
$config['sess_expiration'] = 7200;
$config['sess_save_path'] = sys_get_temp_dir();
$config['encryption_key'] = 'your-secret-key-32-chars';
```

```go
// GoIgniter: Initialize once at application startup (main.go)
package main

import (
    "net/http"
    "github.com/semutdev/goigniter/system/libraries/session"
)

func initSession() {
    session.Init(session.Config{
        Secret:     "your-32-byte-secret-hmac-key!!", // HMAC signing key
        CookieName: "goigniter_session",
        MaxAge:     86400,                            // 24 hours
        Path:       "/",
        HttpOnly:   true,                             // default true
        Secure:     false,                            // set true in production (HTTPS)
        SameSite:   http.SameSiteLaxMode,
        Encrypt:    false,                            // set true to enable AES-256-GCM
        // EncryptKey: []byte("32-bytes-long-key-for-aes256!!"),
    })
}
```

### 6.2 Session Operations

```php
// CI3: Session operations
// Storing data
$this->session->set_userdata('user_id', 123);
$this->session->set_userdata('username', 'john_doe');

// Reading data
$user_id = $this->session->userdata('user_id');
$username = $this->session->userdata('username');

// Removing data
$this->session->unset_userdata('username');

// Destroying entire session
$this->session->sess_destroy();
```

```go
// GoIgniter: Session operations
package controllers

import (
    "github.com/semutdev/goigniter/system/core"
    "github.com/semutdev/goigniter/system/libraries/session"
)

func (u *UserController) Login(c *core.Context) {
    sess := session.Get(c)

    // Storing data
    sess.Set("user_id", 123)
    sess.Set("username", "john_doe")

    // Save session back to cookie
    if err := sess.Save(c); err != nil {
        c.String(500, "Failed to save session")
        return
    }
}

func (u *UserController) Profile(c *core.Context) {
    sess := session.Get(c)

    // Reading data
    userID := sess.GetInt("user_id")
    username := sess.GetString("username")
    rawObj := sess.Get("custom_object")

    // Removing a key
    sess.Delete("temp_key")
    sess.Save(c)
}

func (u *UserController) Logout(c *core.Context) {
    sess := session.Get(c)
    sess.Destroy(c)
    c.Redirect(302, "/login")
}
```

### 6.3 Flashdata

Flash messages are kept only until the subsequent request is completed.

```php
// CI3: Flashdata
$this->session->set_flashdata('success', 'Profile updated successfully!');
$message = $this->session->flashdata('success');
```

```go
// GoIgniter: Flashdata API
import "github.com/semutdev/goigniter/system/libraries/session"

// Setting flashdata (automatically saves session cookie)
session.SetFlash(c, "success", "Profile updated successfully!")

// Reading flashdata (automatically consumes & removes key)
message := session.GetFlash(c, "success")

// Checking whether flash message exists
if session.HasFlash(c, "success") {
    // ...
}
```

---

## 7. Database & Query Builder (Active Record)

GoIgniter provides an Active Record-inspired Query Builder in `github.com/semutdev/goigniter/system/libraries/database` with drivers for MySQL (`go-sql-driver/mysql`) and zero-CGO SQLite (`modernc.org/sqlite`).

### 7.1 Connection Setup

```php
// CI3: application/config/database.php
$db['default'] = array(
    'dsn'      => '',
    'hostname' => 'localhost',
    'username' => 'root',
    'password' => 'secret',
    'database' => 'myapp',
    'dbdriver' => 'mysqli',
);
```

```go
// GoIgniter: database.Open & SetDefault
package main

import (
    "log"
    "github.com/semutdev/goigniter/system/libraries/database"
)

func initDB() {
    // MySQL connection
    dsn := "root:secret@tcp(127.0.0.1:3306)/myapp?parseTime=true&charset=utf8mb4"
    db, err := database.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Set as global default builder instance
    database.SetDefault(db)

    // SQLite alternative:
    // db, err := database.Open("sqlite", "./data/app.db")
}
```

### 7.2 Query Builder API Translation

#### SELECT & FROM
```php
// CI3
$this->db->select('id, name, email');
$query = $this->db->get('users');
$users = $query->result();
```
```go
// GoIgniter
var users []User
err := database.Table("users").
    Select("id", "name", "email").
    Get(&users)
```

#### WHERE Clauses
```php
// CI3
$this->db->where('id', 1);
$this->db->where('age >=', 18);
$this->db->or_where('status', 'active');
$this->db->where_in('role_id', [1, 2, 3]);
$this->db->where_not_in('status', ['banned', 'deleted']);
$this->db->where('deleted_at IS NULL');
$this->db->where('deleted_at IS NOT NULL');
$this->db->where("DATE(created_at) = '2026-01-01'");
```
```go
// GoIgniter
builder := database.Table("users").
    Where("id", 1).
    Where("age", ">=", 18).
    OrWhere("status", "active").
    WhereIn("role_id", []int{1, 2, 3}).
    WhereNotIn("status", []string{"banned", "deleted"}).
    WhereNull("deleted_at").
    WhereNotNull("deleted_at").
    WhereRaw("DATE(created_at) = ?", "2026-01-01")
```

#### JOINs
```php
// CI3
$this->db->select('users.*, roles.name as role_name');
$this->db->from('users');
$this->db->join('roles', 'roles.id = users.role_id');
$this->db->join('profiles', 'profiles.user_id = users.id', 'left');
$this->db->join('teams', 'teams.id = users.team_id', 'right');
$users = $this->db->get()->result();
```
```go
// GoIgniter
var users []UserDTO
err := database.Table("users").
    Select("users.*", "roles.name as role_name").
    Join("roles", "roles.id", "=", "users.role_id").
    LeftJoin("profiles", "profiles.user_id", "=", "users.id").
    RightJoin("teams", "teams.id", "=", "users.team_id").
    Get(&users)
```

#### ORDER BY, LIMIT, OFFSET, GROUP BY & HAVING
```php
// CI3
$this->db->order_by('created_at', 'DESC');
$this->db->limit(10, 20); // limit 10, offset 20
$this->db->group_by('department_id');
$this->db->having('COUNT(*) >', 5);
```
```go
// GoIgniter
builder := database.Table("users").
    OrderBy("created_at", "DESC").
    Limit(10).
    Offset(20).
    GroupBy("department_id").
    Having("COUNT(*) > ?", 5)
```

#### Fetching Single Record (row / row_array)
```php
// CI3
$user = $this->db->where('id', $id)->get('users')->row(); // as object
$user = $this->db->where('id', $id)->get('users')->row_array(); // as array
```
```go
// GoIgniter
var user User
err := database.Table("users").Where("id", id).First(&user)

// Or as a map[string]any:
userMap, err := database.Table("users").Where("id", id).FirstMap()
```

#### Fetching Multiple Records (result / result_array)
```php
// CI3
$users = $this->db->where('status', 'active')->get('users')->result(); // objects
$users = $this->db->where('status', 'active')->get('users')->result_array(); // arrays
```
```go
// GoIgniter
var users []User
err := database.Table("users").Where("status", "active").Get(&users)

// Or as a slice of maps ([]map[string]any):
userMaps, err := database.Table("users").Where("status", "active").GetMap()
```

#### Aggregates
```php
// CI3
$total = $this->db->where('active', 1)->count_all_results('users');
$this->db->select_sum('amount');
$this->db->select_avg('score');
$this->db->select_min('price');
$this->db->select_max('price');
```
```go
// GoIgniter
count, err := database.Table("users").Where("active", 1).Count()
sum, err   := database.Table("orders").Sum("amount")
avg, err   := database.Table("reviews").Avg("score")
min, err   := database.Table("products").Min("price")
max, err   := database.Table("products").Max("price")
```

#### INSERT
```php
// CI3
$data = array(
    'username' => 'alice',
    'email'    => 'alice@example.com',
    'role_id'  => 2
);
$this->db->insert('users', $data);
$new_id = $this->db->insert_id();
```
```go
// GoIgniter: Insert via map
data := map[string]any{
    "username": "alice",
    "email":    "alice@example.com",
    "role_id":  2,
}
err := database.Table("users").Insert(data)

// Or insert and retrieve last insert ID in one call:
newID, err := database.Table("users").InsertGetId(data)

// Or insert directly from struct (uses db:"..." tags):
user := User{Username: "alice", Email: "alice@example.com", RoleID: 2}
err = database.Table("users").InsertStruct(user)
newID, err = database.Table("users").InsertStructGetId(user)
```

#### UPDATE
```php
// CI3
$data = array('email' => 'alice_new@example.com');
$this->db->where('id', 1);
$this->db->update('users', $data);
```
```go
// GoIgniter
err := database.Table("users").
    Where("id", 1).
    Update(map[string]any{"email": "alice_new@example.com"})

// Or update from struct:
err = database.Table("users").Where("id", 1).UpdateStruct(user)
```

#### DELETE
```php
// CI3
$this->db->where('id', 1);
$this->db->delete('users');
```
```go
// GoIgniter
err := database.Table("users").Where("id", 1).Delete()
```

#### Raw Queries
```php
// CI3
$query = $this->db->query("SELECT * FROM users WHERE status = ? AND age > ?", ['active', 21]);
$users = $query->result();

$this->db->query("UPDATE users SET points = points + 10 WHERE id = ?", [1]);
```
```go
// GoIgniter
// Raw SELECT scanning into structs
var users []User
err := database.Query("SELECT * FROM users WHERE status = ? AND age > ?", "active", 21).Get(&users)

// Raw SELECT first record
var single User
err = database.Query("SELECT * FROM users WHERE id = ?", 1).First(&single)

// Raw write execution (INSERT / UPDATE / DELETE)
res, err := database.Exec("UPDATE users SET points = points + 10 WHERE id = ?", 1)
rowsAffected, _ := res.RowsAffected()
```

#### Debugging Queries
```php
// CI3
echo $this->db->last_query();
```
```go
// GoIgniter
sql := database.Table("users").Where("status", "active").Limit(5).ToSQL()
fmt.Println(sql) // Prints SQL query string
```

#### Transactions
```php
// CI3: Manual or automatic transactions
$this->db->trans_start();
$this->db->insert('orders', $order_data);
$this->db->where('id', $product_id)->update('products', ['stock' => $new_stock]);
$this->db->trans_complete();

if ($this->db->trans_status() === FALSE) {
    // Transaction failed
}
```
```go
// GoIgniter: Idiomatic closure-based transaction (auto commit/rollback on error)
err := database.Transaction(func(tx *database.DB) error {
    if err := tx.Table("orders").Insert(orderData); err != nil {
        return err // Triggers automatic rollback
    }
    if err := tx.Table("products").Where("id", productID).Update(stockData); err != nil {
        return err // Triggers automatic rollback
    }
    return nil // Commits transaction
})
```

---

## 8. Views & Layout Rendering

GoIgniter uses Go's standard `html/template` with hot-reload capabilities and customizable template functions.

### 8.1 Template Engine Initialization

```go
// Initialize in main.go
isDev := os.Getenv("APP_ENV") != "production"

// Load templates with reload flag (true enables auto-reloading on every request)
app.LoadTemplatesWithFuncs("views", isDev, helpers.AllTemplateFuncs())
```

### 8.2 Rendering Views from Controller

```php
// CI3: application/controllers/User.php
$data['title'] = 'User List';
$data['users'] = $users;
$this->load->view('users/index', $data);
```

```go
// GoIgniter: Controller action
c.View("users/index", core.Map{
    "Title": "User List",
    "Users": users,
})

// Or inside a struct method inheriting core.Controller:
u.Ctx.View("users/index", core.Map{
    "Title": "User List",
    "Users": users,
})
```

### 8.3 Layout Composition

```php
// CI3: Typical multi-view layout pattern
$this->load->view('layouts/header', $data);
$this->load->view('users/index', $data);
$this->load->view('layouts/footer', $data);
```

```go
// GoIgniter: Partial string rendering composition pattern
content, err := c.Render("users/index", core.Map{"Users": users})
if err != nil {
    return c.String(500, err.Error())
}

return c.View("layouts/main", core.Map{
    "Title":   "User List",
    "Content": template.HTML(content), // or safe helper in template
})
```

### 8.4 PHP View to Go HTML Template Syntax Mapping

| Template Construct | CodeIgniter 3 PHP Template | GoIgniter `html/template` |
|---|---|---|
| **Echo variable** | `<?= $title ?>` | `{{ .Title }}` |
| **Object property** | `<?= $user->email ?>` | `{{ .User.Email }}` |
| **Array key** | `<?= $user['email'] ?>` | `{{ .User.Email }}` |
| **If condition** | `<?php if ($logged_in): ?>...<?php endif; ?>` | `{{ if .LoggedIn }}...{{ end }}` |
| **If / Else** | `<?php if ($role == 'admin'): ?>...<?php else: ?>...<?php endif; ?>` | `{{ if eq .Role "admin" }}...{{ else }}...{{ end }}` |
| **Foreach loop** | `<?php foreach ($users as $u): ?>...<?php endforeach; ?>` | `{{ range $u := .Users }}...{{ end }}` |
| **Empty loop fallback**| `<?php if (empty($users)): ?>...<?php endif; ?>` | `{{ range .Users }}{{ else }}No users found{{ end }}` |
| **Raw HTML output** | `<?= $html_content ?>` | `{{ .HTMLContent \| safe }}` |
| **Base URL** | `<?= base_url('css/app.css') ?>` | `{{ base_url "/css/app.css" }}` |
| **Asset URL** | `<?= base_url('assets/img/logo.png') ?>` | `{{ asset_url "img/logo.png" }}` |
| **String helper** | `<?= strtoupper($name) ?>` | `{{ .Name \| upper }}` |
| **Default value** | `<?= $bio ?: 'No bio provided' ?>` | `{{ default "No bio provided" .Bio }}` |

---

## 9. File Upload & Image Manipulation

GoIgniter provides a robust upload library (`github.com/semutdev/goigniter/system/libraries/upload`) offering strict MIME verification, extension validation, path traversal prevention, and image processing.

### 9.1 Upload Configuration & Execution

```php
// CI3: File upload
$config['upload_path']   = './uploads/';
$config['allowed_types'] = 'gif|jpg|png|pdf';
$config['max_size']      = 2048; // in KB
$config['encrypt_name']  = TRUE;

$this->load->library('upload', $config);

if (!$this->upload->do_upload('userfile')) {
    $error = $this->upload->display_errors();
    // handle error
} else {
    $data = $this->upload->data();
    $saved_filename = $data['file_name'];
    $full_path = $data['full_path'];
}
```

```go
// GoIgniter: File upload
package controllers

import (
    "net/http"
    "github.com/semutdev/goigniter/system/core"
    "github.com/semutdev/goigniter/system/libraries/upload"
)

func (u *UserController) DoUpload(c *core.Context) {
    config := upload.Config{
        UploadPath:   "./public/uploads",
        AllowedTypes: "gif|jpg|jpeg|png|pdf",
        MaxSize:      2048,          // Maximum size in KB
        FileName:     "timestamp",   // "timestamp", "random", "original", or custom
        Overwrite:    false,
        CreateDirs:   true,          // Auto-create directory if missing
    }

    uploader := upload.New(config)
    result, err := uploader.Do("userfile", c.Request)
    if err != nil {
        switch err {
        case upload.ErrNoFile:
            c.JSON(http.StatusBadRequest, core.Map{"error": "No file uploaded"})
        case upload.ErrFileTooBig:
            c.JSON(http.StatusBadRequest, core.Map{"error": "File size exceeds limit"})
        case upload.ErrInvalidType:
            c.JSON(http.StatusBadRequest, core.Map{"error": "File type not permitted"})
        case upload.ErrSuspiciousFile, upload.ErrMimeMismatch:
            c.JSON(http.StatusBadRequest, core.Map{"error": "Security validation failed"})
        default:
            c.JSON(http.StatusInternalServerError, core.Map{"error": err.Error()})
        }
        return
    }

    // Access uploaded file metadata
    c.JSON(http.StatusOK, core.Map{
        "file_name":     result.FileName,
        "original_name": result.OriginalName,
        "file_path":     result.FilePath,
        "file_size":     result.FileSize,
        "file_type":     result.FileType,
        "is_image":      result.IsImage,
        "image_width":   result.ImageWidth,
        "image_height":  result.ImageHeight,
    })
}
```

### 9.2 Low-Level Native Upload

If custom multipart streaming or manual processing is preferred, standard Go stdlib methods work directly on `c.Request`:

```go
file, header, err := c.Request.FormFile("userfile")
if err != nil {
    // handle error
}
defer file.Close()
```

### 9.3 Image Manipulation

```php
// CI3: Image resize via image_lib
$config['image_library'] = 'gd2';
$config['source_image'] = './uploads/photo.jpg';
$config['maintain_ratio'] = TRUE;
$config['width'] = 800;
$config['height'] = 600;

$this->load->library('image_lib', $config);
$this->image_lib->resize();
```

```go
// GoIgniter: Image processor
imgProcessor := upload.NewImageProcessor(upload.ImageConfig{
    Source:              result.FilePath,
    Destination:         "./public/uploads/thumbs/" + result.FileName,
    Width:               800,
    Height:              600,
    MaintainAspectRatio: true,
    Quality:             85,
})

if err := imgProcessor.Resize(); err != nil {
    // handle image processing error
}

// Additional operations supported:
// imgProcessor.Crop(width, height, x, y)
// imgProcessor.Thumbnail(150, 150)
// imgProcessor.Rotate(90)
```

---

## 10. Middleware & Security

CodeIgniter 3 handles cross-cutting concerns via system hooks (`application/config/hooks.php`) or inline checks. GoIgniter provides composable, high-performance middleware in `system/middleware`.

### 10.1 Global Middleware Pipeline

```go
// Register in main.go
package main

import (
    "time"
    "github.com/semutdev/goigniter/system/core"
    "github.com/semutdev/goigniter/system/middleware"
)

func main() {
    app := core.New()

    // 1. Recovery from panics (500 internal server error fallback)
    app.Use(middleware.Recovery())

    // 2. HTTP Request Logger
    app.Use(middleware.Logger())

    // 3. Security Headers (X-XSS-Protection, nosniff, Frame Options)
    app.Use(middleware.SecurityHeaders())

    // 4. CORS
    app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"https://example.com"},
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
    }))

    // 5. Rate Limiting (e.g. 100 requests per minute per IP)
    app.Use(middleware.RateLimit(100, time.Minute))

    // 6. CSRF Protection
    app.Use(middleware.CSRF())

    app.Run(":8080")
}
```

### 10.2 CSRF Protection Mapping

In CI3, CSRF is toggled in `config.php`:
```php
// CI3 config.php
$config['csrf_protection'] = TRUE;
$config['csrf_token_name'] = 'csrf_token';
$config['csrf_cookie_name'] = 'csrf_cookie';
```

In GoIgniter:
```go
// GoIgniter: Apply CSRF middleware
app.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
    TokenLength:   32,
    CookieName:    "csrf_token",
    HeaderName:    "X-CSRF-Token",
    FormFieldName: "csrf_token",
    Secure:        false, // true for HTTPS
}))
```

In HTML templates:
```html
<!-- CI3 -->
<input type="hidden" name="<?= $this->security->get_csrf_token_name(); ?>" value="<?= $this->security->get_csrf_hash(); ?>">

<!-- GoIgniter -->
<input type="hidden" name="csrf_token" value="{{ .csrf_token }}">
```

For AJAX / Fetch requests:
```javascript
// Header sent with AJAX:
// Header Name: X-CSRF-Token
// Value: Read from "csrf_token" cookie
```

### 10.3 Authentication Middleware

```go
// Basic Auth
app.Use(middleware.BasicAuth(func(user, pass string) bool {
    return user == "admin" && pass == "secret"
}))

// Bearer Token Auth
apiGroup := app.Group("/api", middleware.BearerAuth(func(token string) bool {
    return token == "valid-api-secret-key"
}))
```

---

## 11. Helpers & Common Utilities

| CI3 Helper / Function | GoIgniter Helper Function | Package / Usage |
|---|---|---|
| `base_url($path)` | `helpers.BaseURL(path)` | `"github.com/semutdev/goigniter/system/helpers"` |
| `site_url($path)` | `helpers.SiteURL(path)` | Alias to `BaseURL` |
| `base_url('public/'.$path)` | `helpers.AssetURL(path)` | Prefixes `/public/` |
| `current_url()` | `c.Request.URL.String()` | Context request URL |
| `var_dump($var)` | `helpers.Dump(var)` | Pretty printer |
| `json_encode($data)` | `helpers.DumpJSON(data)` | Formatted JSON dumper |
| `redirect($url)` | `c.Redirect(http.StatusFound, url)` | HTTP 302 redirect |
| `show_404()` | `c.HTML(404, "Page Not Found")` | 404 response |
| `show_error($msg, 500)` | `c.String(500, msg)` | 500 response |

---

## 12. Comprehensive CI3 vs GoIgniter API Cheat Table

| Category | CodeIgniter 3 (PHP) | GoIgniter (Go) | Notes |
|---|---|---|---|
| **Route GET** | `$route['users'] = 'user/index';` | `app.GET("/users", handler)` | Static route |
| **Route Param** | `$route['users/(:num)'] = 'user/show/$1';` | `app.GET("/users/:id", handler)` | Captured via `c.Param("id")` |
| **Route Wildcard**| `$route['files/(:any)'] = 'files/get/$1';` | `app.GET("/files/*path", handler)` | Captured via `c.Param("path")` |
| **Form POST** | `$this->input->post('name')` | `c.Form("name")` | Empty string if not present |
| **Query GET** | `$this->input->get('page')` | `c.Query("page")` | `c.QueryDefault("page", "1")` |
| **Query Int** | `(int)$this->input->get('page')` | `c.QueryInt("page")` | Returns `(int, error)` |
| **Route Param** | `$this->uri->segment(3)` | `c.Param("id")` | By name instead of 1-based index |
| **Route Param Int**| `(int)$this->uri->segment(3)` | `c.ParamInt("id")` | Returns `(int, error)` |
| **Header Get** | `$this->input->get_request_header('X')` | `c.Header("X")` | Case-insensitive |
| **Header Set** | `$this->output->set_header('X: 1')` | `c.SetHeader("X", "1")` | Sets response header |
| **JSON Output** | `echo json_encode($data);` | `c.JSON(200, data)` | Auto sets content-type |
| **Text Output** | `echo $text;` | `c.String(200, text)` | Plain text |
| **HTML Output** | `$this->output->set_output($h);` | `c.HTML(200, html)` | Raw HTML |
| **Render View** | `$this->load->view('name', $d);` | `c.View("name", data)` | Renders html/template |
| **Redirect** | `redirect('/login');` | `c.Redirect(302, "/login")` | StatusFound (302) |
| **Session Read** | `$this->session->userdata('uid')` | `sess.Get("uid")` / `sess.GetInt("uid")` | Type-specific helpers available |
| **Session Write**| `$this->session->set_userdata('k', $v)` | `sess.Set("k", v); sess.Save(c)` | Explicit `Save(c)` |
| **Session Del** | `$this->session->unset_userdata('k')` | `sess.Delete("k"); sess.Save(c)` | Remove item |
| **Session Kill** | `$this->session->sess_destroy()` | `sess.Destroy(c)` | Clears session cookie |
| **Flash Set** | `$this->session->set_flashdata('m', $v)`| `session.SetFlash(c, "m", v)` | Auto-saves cookie |
| **Flash Get** | `$this->session->flashdata('m')` | `session.GetFlash(c, "m")` | Consumes & clears flash |
| **DB Select** | `$this->db->select('a, b')` | `db.Table("t").Select("a", "b")` | Column projection |
| **DB Where** | `$this->db->where('id', 1)` | `db.Table("t").Where("id", 1)` | Operator defaults to `=` |
| **DB Or Where** | `$this->db->or_where('id', 2)` | `db.Table("t").OrWhere("id", 2)` | Appends `OR` condition |
| **DB Where In** | `$this->db->where_in('id', $ids)` | `db.Table("t").WhereIn("id", ids)` | Accepts `[]int`, `[]string`, etc. |
| **DB Where Null**| `$this->db->where('col IS NULL')` | `db.Table("t").WhereNull("col")` | IS NULL |
| **DB Order By** | `$this->db->order_by('col', 'DESC')` | `db.Table("t").OrderBy("col", "DESC")` | Column + direction |
| **DB Limit** | `$this->db->limit(10, 5)` | `db.Table("t").Limit(10).Offset(5)` | Separate limit and offset |
| **DB Get All** | `$query = $this->db->get('users')` | `db.Table("users").Get(&users)` | Scans into slice of structs |
| **DB Get One** | `$row = $this->db->get('users')->row()`| `db.Table("users").First(&user)` | Scans into single struct |
| **DB Count** | `$this->db->count_all_results('u')` | `count, err := db.Table("u").Count()`| Returns `int64` |
| **DB Insert** | `$this->db->insert('u', $data)` | `db.Table("u").Insert(map)` | Or `InsertStruct(model)` |
| **DB Insert ID**| `$this->db->insert_id()` | `id, _ := db.Table("u").InsertGetId(d)`| Returns `(int64, error)` |
| **DB Update** | `$this->db->update('u', $data)` | `db.Table("u").Where(...).Update(d)` | Or `UpdateStruct(model)` |
| **DB Delete** | `$this->db->delete('u')` | `db.Table("u").Where(...).Delete()` | Execute delete |
| **DB Raw Query**| `$this->db->query($sql, $binds)` | `db.Query(sql, binds...).Get(&dest)` | Raw query scanning |
| **DB Exec** | `$this->db->query($raw_update)` | `db.Exec(sql, binds...)` | Raw DML execution |
| **Transaction** | `$this->db->trans_start(); ...` | `db.Transaction(func(tx) error { ... })`| Automatic rollback on error |
| **Upload Init** | `$this->load->library('upload', $cfg)` | `uploader := upload.New(cfg)` | Struct config |
| **Upload Do** | `$this->upload->do_upload('file')` | `result, err := uploader.Do("file", c.Request)` | Returns typed `*Result` |
| **CSRF Token** | `$this->security->get_csrf_hash()` | `{{ .csrf_token }}` / Header `X-CSRF-Token`| Constant-time token check |

---

## 13. Common Migration Gotchas & Best Practices

1. **Context Lifecycle & Goroutines**:
   `*core.Context` instances are recycled via `sync.Pool`. Never retain a reference to `c` inside a background goroutine without copying needed values (e.g. string parameters, user IDs) beforehand.

2. **Form Validation Strategy**:
   GoIgniter does not implement CI3's string-based `$this->form_validation->set_rules()`. Idiomatic Go validates structs directly inside model or request structs using standard type checking or validation libraries (e.g., `go-playground/validator`).

3. **Database Nullable Fields**:
   In CI3, PHP sets `null` dynamically. In Go, database columns that can be `NULL` should be modeled using pointers (`*string`, `*int`, `*time.Time`) or `sql.NullString` / `sql.NullInt64` to prevent scanner errors.

4. **Template Variables Title Casing**:
   Go templates cannot access unexported struct fields or lowercase keys directly if passed as structs. Keep struct fields exported (uppercase first letter) and use `{{ .UserName }}` rather than `{{ .username }}`.

5. **Static Assets Location**:
   CI3 projects often mix static assets in the root folder. In GoIgniter, place all static CSS, JS, and image assets in `./public` and serve via `app.Static("/assets", "./public/assets")`.
