package libs

// MethodRestrictor interface untuk membatasi HTTP method pada controller
// Implement interface ini jika ingin restrict method tertentu (untuk API)
type MethodRestrictor interface {
	AllowedMethods() map[string][]string
}
