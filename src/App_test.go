package main

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestRunApp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		f, tmpFileErr := os.CreateTemp("", "db_*.db")
		assert.NoError(t, tmpFileErr)
		defer os.Remove(f.Name())

		_ = os.Setenv("DATABASE_FILEPATH", f.Name())
		defer os.Unsetenv("DATABASE_FILEPATH")

		var appErr error
		go func() {
			appErr = RunApp()
		}()
		runtime.Gosched()

		var err error
		var res *http.Response
		for i := 0; i < 3; i++ {
			if appErr != nil {
				t.Errorf("RunApp() error = %v", appErr)
				break
			}

			time.Sleep(50 * time.Millisecond)
			client := http.Client{
				Timeout: time.Second * 2,
			}
			res, err = client.Get("http://localhost:8080/healthcheck")
			if err == nil {
				break
			}
		}

		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		assert.Equal(t, "health", string(body))
	})

	t.Run("fail", func(t *testing.T) {
		os.Unsetenv("DATABASE_FILEPATH")

		var err error
		go func() {
			err = RunApp()
		}()
		runtime.Gosched()
		if err == nil {
			time.Sleep(50 * time.Millisecond)
		}
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})
}

func TestHandleExitError(t *testing.T) {
	t.Run("Handle exit error", func(t *testing.T) {
		var actualExitCode int
		var out bytes.Buffer

		testCases := map[error]int{
			errors.New("dummy error"): ExitCodeMainError,
			nil:                       0,
		}

		for err, expectedCode := range testCases {
			out.Reset()
			actualExitCode = HandleExitError(&out, err)

			assert.Equal(t, expectedCode, actualExitCode)
			if err == nil {
				assert.Empty(t, out.String(), "Error is not empty")
			} else {
				assert.Contains(t, out.String(), err.Error(), "error output hasn't error description")
			}
		}
	})
}
