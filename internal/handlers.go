package internal

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"users/pkg/database"
	"users/pkg/url"
)

type UserAPI struct {
	DB      database.Database
	Manager IUserManager
}

// Login endpoint
func (a *UserAPI) Login(c echo.Context) error {
	userReq := &User{}
	if err := c.Bind(userReq); err != nil {
		return NewErrorResponse(c, ErrWrongBody)
	}
	user, err := a.Manager.Login(*userReq)
	if err != nil {
		return NewErrorResponse(c, err)
	}
	cleanUser(user)
	return c.JSON(http.StatusOK, user)
}

// GetUserHandler endpoint to get user information
func (a *UserAPI) GetUserHandler(c echo.Context) error {
	var ID string
	if err := url.ParseURLPath(c, url.PathMap{
		ParamUserID: {Target: &ID, Err: ErrUserIDNotPresent},
	}); err != nil {
		return NewErrorResponse(c, err)
	}

	user, err := a.Manager.GetUser(ID)
	if err != nil {
		return NewErrorResponse(c, err)
	}
	cleanUser(user)
	return c.JSON(http.StatusOK, user)
}

// PostUserHandler endpoint to create a new user
func (a *UserAPI) PostUserHandler(c echo.Context) error {
	userReq := &User{}
	if err := c.Bind(userReq); err != nil {
		return NewErrorResponse(c, ErrWrongBody)
	}

	user, err := a.Manager.CreateUser(*userReq)
	if err != nil {
		return NewErrorResponse(c, err)
	}
	cleanUser(user)
	return c.JSON(http.StatusCreated, user)
}

// PutUserHandler endpoint to update an existing user
func (a *UserAPI) PutUserHandler(c echo.Context) error {
	var ID string
	if err := url.ParseURLPath(c, url.PathMap{
		ParamUserID: {Target: &ID, Err: ErrUserIDNotPresent},
	}); err != nil {
		return NewErrorResponse(c, err)
	}

	userReq := &User{}
	if err := c.Bind(userReq); err != nil {
		return NewErrorResponse(c, ErrWrongBody)
	}

	user, err := a.Manager.UpdateUser(ID, *userReq)
	if err != nil {
		return NewErrorResponse(c, err)
	}
	cleanUser(user)
	return c.JSON(http.StatusOK, user)
}

// DeleteUserHandler endpoint to delete an existing user
func (a *UserAPI) DeleteUserHandler(c echo.Context) error {
	var ID string
	if err := url.ParseURLPath(c, url.PathMap{
		ParamUserID: {Target: &ID, Err: ErrUserIDNotPresent},
	}); err != nil {
		return NewErrorResponse(c, err)
	}

	err := a.Manager.DeleteUser(ID)
	if err != nil {
		return NewErrorResponse(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func cleanUser(user *User) *User {
	user.Password = ""
	return user
}
