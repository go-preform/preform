package mainModel
type PreformTestALogType string

var (
	PreformTestALogTypes = struct{
		Register PreformTestALogType
		Login PreformTestALogType
	}{
		Register: "Register",
		Login: "Login",
	}
)
