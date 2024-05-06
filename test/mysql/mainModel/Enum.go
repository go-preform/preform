package mainModel
type PreformTestAUserLogType string

var (
	PreformTestAUserLogTypes = struct{
		Register PreformTestAUserLogType
		Login PreformTestAUserLogType
	}{
		Register: "Register",
		Login: "login",
	}
)
