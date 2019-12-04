package util

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	_ = os.Setenv("TEST_VAR", "TEST_VAL")
	actual := GetEnv("TEST_VAR", "OOPS")
	if actual != "TEST_VAL" {
		t.Errorf("start failed, expected %s, got %s", "TEST_VAL", actual)
	}
}

func TestGetEnvAsInt(t *testing.T) {
	_ = os.Setenv("TEST_VAR", "123")
	actual := GetEnvAsInt("TEST_VAR", 321)
	if actual != 123 {
		t.Errorf("start failed, expected %d, got %d", 123, actual)
	}
}

func TestIsASCII(t *testing.T) {
	require.True(t, IsASCII("Hello"))
	require.False(t, IsASCII("Привет"))
}

func TestFileExists(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "util_test")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		f.Close()
	}()

	require.True(t, FileExists(f.Name()))
}

func TestIsBlank(t *testing.T) {
	require.True(t, IsBlank(""))
	require.True(t, IsBlank("   "))
	require.False(t, IsBlank(" test  "))
	require.False(t, IsBlank("test"))
}
