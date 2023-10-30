package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
)

const ExitCodeMainError = 1

func RunApp() error {
	gin.SetMode(gin.ReleaseMode)

	serviceContainer, err := BuildServiceContainer(os.Getenv("DATABASE_FILEPATH"))

	if err == nil {
		defer serviceContainer.Database.Close()
		err = http.ListenAndServe(":8080", serviceContainer.Router)
	}

	return err
}

func HandleExitError(errStream io.Writer, err error) int {
	if err != nil {
		_, _ = fmt.Fprintln(errStream, err)
	}

	if err != nil {
		return ExitCodeMainError
	}

	return 0
}
