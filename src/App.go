package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
)

const ExitCodeMainError = 1

const ListenPort = ":8080"

func RunApp() error {
	gin.SetMode(gin.ReleaseMode)

	serviceContainer, err := BuildServiceContainer(os.Getenv("DATABASE_FILEPATH"))

	if err == nil {
		serviceContainer.WebhookDispatcher.Start()
		defer serviceContainer.WebhookDispatcher.Close()
		defer serviceContainer.Database.Close()

		err = http.ListenAndServe(ListenPort, serviceContainer.Router)
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
