package libs

var controllerRegistry = make(map[string]interface{})

// Register mendaftarkan controller ke registry global
// Dipanggil di init() setiap controller
func Register(name string, controller interface{}) {
	controllerRegistry[name] = controller
}

// GetRegistry mengembalikan semua controller yang terdaftar
func GetRegistry() map[string]interface{} {
	return controllerRegistry
}
