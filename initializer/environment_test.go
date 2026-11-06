package initializer

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoadEnvironmentVariables(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Cleanup(func() { gin.SetMode(gin.TestMode) })

	// t.Setenv registers restoration; godotenv.Load does not override existing vars.
	t.Setenv("GIN_MODE", "test")
	t.Cleanup(func() { os.Unsetenv("INITIALIZER_TEST_VAR") })

	if err := os.WriteFile(".env", []byte("INITIALIZER_TEST_VAR=hello\nGIN_MODE=release\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	LoadEnvironmentVariables()

	if got := os.Getenv("INITIALIZER_TEST_VAR"); got != "hello" {
		t.Errorf("INITIALIZER_TEST_VAR = %q, want hello", got)
	}
	if gin.Mode() != gin.TestMode {
		t.Errorf("gin mode = %q, want %q", gin.Mode(), gin.TestMode)
	}
}
