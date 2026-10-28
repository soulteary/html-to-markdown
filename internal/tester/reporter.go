package tester

// reporter allows failure diagnostics to be checked without deliberately
// failing the parent test. *testing.T implements it.
type reporter interface {
	Helper()
	Error(...any)
	Errorf(string, ...any)
	Fatal(...any)
	Fatalf(string, ...any)
	Logf(string, ...any)
}
