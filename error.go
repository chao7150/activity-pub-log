package activitypublog

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HandlerError(method string, path string, c echo.Context) func(error) error {
	return func(err error) error {
		errString := fmt.Sprintf("error %s %s: %v", method, path, err)
		fmt.Println(errString)
		return c.String(http.StatusInternalServerError, errString)
	}
}
