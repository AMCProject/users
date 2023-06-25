package internal

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

const (
	RouteLogin  = "/login"
	RouteUser   = "/user"
	RouteUserID = "/user/:id"

	ParamUserID = "id"
)

type ErrorResponse struct {
	Err ErrorBody `json:"error"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Err.Status, e.Err.Message)
}

type ErrorBody struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func NewErrorResponse(c echo.Context, err error) error {
	errResponse := &ErrorResponse{Err: errorsMap[err.Error()]}
	if err := c.JSON(errResponse.Err.Status, errResponse); err != nil {
		return err
	}
	return errResponse
}

var errorsMap = map[string]ErrorBody{
	ErrUserIDNotPresent.Error():   {Status: http.StatusBadRequest, Message: ErrUserIDNotPresent.Error()},
	ErrWrongBody.Error():          {Status: http.StatusBadRequest, Message: ErrWrongBody.Error()},
	ErrHashingPassword.Error():    {Status: http.StatusBadRequest, Message: ErrHashingPassword.Error()},
	ErrWrongPassword.Error():      {Status: http.StatusBadRequest, Message: ErrWrongPassword.Error()},
	ErrUserNotFound.Error():       {Status: http.StatusNotFound, Message: ErrUserNotFound.Error()},
	ErrUserAlreadyExists.Error():  {Status: http.StatusConflict, Message: ErrUserAlreadyExists.Error()},
	ErrSomethingWentWrong.Error(): {Status: http.StatusInternalServerError, Message: ErrSomethingWentWrong.Error()},
}
var (
	ErrUserIDNotPresent   = errors.New("error con el ID del usuario dado")
	ErrSomethingWentWrong = errors.New("error inesperado")
	ErrUserAlreadyExists  = errors.New("este usuario ya existe")
	ErrWrongBody          = errors.New("el cuerpo enviado es erróneo")
	ErrUserNotFound       = errors.New("usuario no encontrado")
	ErrHashingPassword    = errors.New("error encriptando la contraseña")
	ErrWrongPassword      = errors.New("contraseña errónea")
)
