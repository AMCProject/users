package handlers

import (
	"bytes"
	"github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
	"users/internal"
	"users/internal/managers"
	"users/internal/models"
	"users/pkg/database"
)

var databaseTest = "/amc_test.db"

type UserAPITestSuite struct {
	suite.Suite
	db *database.Database
}

func TestUserAPITestSuite(t *testing.T) {
	suite.Run(t, new(UserAPITestSuite))
}

func (s *UserAPITestSuite) SetupTest() {
	_ = database.RemoveDB(databaseTest)
	s.db = database.InitDB(databaseTest)
	password, _ := managers.HashPassword("MyPassword.123")
	s.db.Conn.Exec("INSERT INTO users(id,name,mail,password) VALUES (?,?,?,?)", "01FN3EEB2NVFJAHAPU00000001", "firstuser", "firstuser@mail.com", password)
}

func (s *UserAPITestSuite) TearDownTest() {
	s.db = nil
	_ = database.RemoveDB(databaseTest)
}
func (s *UserAPITestSuite) TestLoginHandler() {
	tests := []struct {
		name               string
		reqBody            interface{}
		expectedResp       interface{}
		expectedStatusCode int
		wantErr            bool
	}{
		{
			name: "[001] Login user (ok)",
			reqBody: &models.User{
				Mail:     "firstuser@mail.com",
				Password: "MyPassword.123",
			},
			expectedResp: &models.User{
				Id:       "01FN3EEB2NVFJAHAPU00000001",
				Name:     PointerString("firstuser"),
				Mail:     "firstuser@mail.com",
				Password: "",
			},
			expectedStatusCode: http.StatusOK,
			wantErr:            false,
		},
		{
			name: "[002] Login user not found (400)",
			reqBody: &models.User{
				Mail:     "inventeduser@mail.com",
				Password: "MyPassword2.123",
			},
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusNotFound,
					Message: internal.ErrUserNotFound.Error(),
				},
			},
			expectedStatusCode: http.StatusNotFound,
			wantErr:            true,
		},
		{
			name: "[003] Wrong password (400)",
			reqBody: &models.User{
				Mail:     "firstuser@mail.com",
				Password: "MyPassword2.123",
			},
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusBadRequest,
					Message: internal.ErrWrongPassword.Error(),
				},
			},
			expectedStatusCode: http.StatusBadRequest,
			wantErr:            true,
		},
		{
			name:    "[004] Wrong struct sent (400)",
			reqBody: "invalid",
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusBadRequest,
					Message: internal.ErrWrongBody.Error(),
				},
			},
			expectedStatusCode: http.StatusBadRequest,
			wantErr:            true,
		},
	}
	getEchoContext := func(request interface{}) echo.Context {
		var body []byte
		body, err := jsoniter.Marshal(request)
		s.NoError(err)
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, internal.RouteLogin, bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		return c
	}
	for _, t := range tests {
		s.Run(t.name, func() {
			userManager := managers.NewUserManager(*s.db)
			api := UserAPI{DB: *s.db, Manager: userManager}

			c := getEchoContext(t.reqBody)
			err := api.Login(c)

			if t.wantErr {
				s.Equal(t.wantErr, err != nil)
				resp, ok := c.Response().Writer.(*httptest.ResponseRecorder)
				s.True(ok)
				body := resp.Body.Bytes()

				errorReturned := new(internal.ErrorResponse)
				s.NoError(jsoniter.Unmarshal(body, errorReturned))
				s.Equal(errorReturned, t.expectedResp)
			} else {
				resp, ok := c.Response().Writer.(*httptest.ResponseRecorder)
				s.True(ok)
				body := resp.Body.Bytes()

				actualUser := new(models.User)
				s.NoError(jsoniter.Unmarshal(body, actualUser))
				s.Equal(actualUser, t.expectedResp)
			}

			s.Equal(t.expectedStatusCode, c.Response().Status)
		})
	}
}

func (s *UserAPITestSuite) TestPostUserHandler() {
	tests := []struct {
		name               string
		reqBody            interface{}
		expectedULID       ulid.ULID
		expectedResp       interface{}
		expectedStatusCode int
		wantErr            bool
	}{
		{
			name: "[001] Create new user (ok)",
			reqBody: &models.User{
				Mail:     "test1@testmail.com",
				Password: "MyPassword.123",
			},
			expectedULID: ulid.MustParse("01FN3EEB2NVFJAHAPVXGDKHXG9"),
			expectedResp: &models.User{
				Id:       "01FN3EEB2NVFJAHAPVXGDKHXG9",
				Name:     PointerString("test1"),
				Mail:     "test1@testmail.com",
				Password: "",
			},
			expectedStatusCode: http.StatusCreated,
			wantErr:            false,
		},
		{
			name: "[002] Create duplicated user (409)",
			reqBody: &models.User{
				Mail:     "test1@testmail.com",
				Password: "MyPassword2.123",
			},
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusConflict,
					Message: internal.ErrUserAlreadyExists.Error(),
				},
			},
			expectedStatusCode: http.StatusConflict,
			wantErr:            true,
		},
		{
			name: "[003] Wrong user struct, mail is missing (400)",
			reqBody: &models.User{
				Password: "MyPassword2.123",
			},
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusBadRequest,
					Message: internal.ErrWrongBody.Error(),
				},
			},
			expectedStatusCode: http.StatusBadRequest,
			wantErr:            true,
		},
		{
			name:    "[004] Wrong struct sent (400)",
			reqBody: "invalid",
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusBadRequest,
					Message: internal.ErrWrongBody.Error(),
				},
			},
			expectedStatusCode: http.StatusBadRequest,
			wantErr:            true,
		},
	}
	getEchoContext := func(request interface{}) echo.Context {
		var body []byte
		body, err := jsoniter.Marshal(request)
		s.NoError(err)
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, internal.RouteUser, bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		return c
	}
	for _, t := range tests {
		s.Run(t.name, func() {
			userManager := managers.NewUserManager(*s.db)
			api := UserAPI{DB: *s.db, Manager: userManager}

			c := getEchoContext(t.reqBody)
			err := api.PostUserHandler(c)

			if t.wantErr {
				s.Equal(t.wantErr, err != nil)
				resp, ok := c.Response().Writer.(*httptest.ResponseRecorder)
				s.True(ok)
				body := resp.Body.Bytes()

				errorReturned := new(internal.ErrorResponse)
				s.NoError(jsoniter.Unmarshal(body, errorReturned))
				s.Equal(errorReturned, t.expectedResp)
			} else {
				resp, ok := c.Response().Writer.(*httptest.ResponseRecorder)
				s.True(ok)
				body := resp.Body.Bytes()

				actualUser := new(models.User)
				s.NoError(jsoniter.Unmarshal(body, actualUser))
				actualUser.Id = t.expectedULID.String()
				s.Equal(actualUser, t.expectedResp)
			}

			s.Equal(t.expectedStatusCode, c.Response().Status)
		})
	}
}

func (s *UserAPITestSuite) TestGetUserHandler() {
	tests := []struct {
		name               string
		userID             string
		expectedResp       interface{}
		expectedStatusCode int
		wantErr            bool
	}{
		{
			name:   "[001] Get user (ok)",
			userID: "01FN3EEB2NVFJAHAPU00000001",
			expectedResp: &models.User{
				Id:       "01FN3EEB2NVFJAHAPU00000001",
				Name:     PointerString("firstuser"),
				Mail:     "firstuser@mail.com",
				Password: "",
			},
			expectedStatusCode: http.StatusOK,
			wantErr:            false,
		},
		{
			name: "[002] Get user, userId not indicated (400)",
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusBadRequest,
					Message: internal.ErrUserIDNotPresent.Error(),
				},
			},
			expectedStatusCode: http.StatusBadRequest,
			wantErr:            true,
		},
		{
			name:   "[003] User does not exist (404)",
			userID: "01FN3EEB2NVFJAHAPU00000099",
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusNotFound,
					Message: internal.ErrUserNotFound.Error(),
				},
			},
			expectedStatusCode: http.StatusNotFound,
			wantErr:            true,
		},
	}
	getEchoContext := func(userId string) echo.Context {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, internal.RouteUserID, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames(internal.ParamUserID)
		c.SetParamValues(userId)
		return c
	}
	for _, t := range tests {
		s.Run(t.name, func() {
			userManager := managers.NewUserManager(*s.db)
			api := UserAPI{DB: *s.db, Manager: userManager}

			c := getEchoContext(t.userID)
			err := api.GetUserHandler(c)

			if t.wantErr {
				s.Equal(t.wantErr, err != nil)
				resp, ok := c.Response().Writer.(*httptest.ResponseRecorder)
				s.True(ok)
				body := resp.Body.Bytes()

				errorReturned := new(internal.ErrorResponse)
				s.NoError(jsoniter.Unmarshal(body, errorReturned))
				s.Equal(errorReturned, t.expectedResp)
			} else {
				resp, ok := c.Response().Writer.(*httptest.ResponseRecorder)
				s.True(ok)
				body := resp.Body.Bytes()

				actualUser := new(models.User)
				s.NoError(jsoniter.Unmarshal(body, actualUser))
				s.Equal(actualUser, t.expectedResp)
			}

			s.Equal(t.expectedStatusCode, c.Response().Status)
		})
	}
}

func (s *UserAPITestSuite) TestPutUserHandler() {
	tests := []struct {
		name               string
		userID             string
		reqBody            interface{}
		expectedResp       interface{}
		expectedStatusCode int
		wantErr            bool
	}{
		{
			name:   "[001] Update user name (ok)",
			userID: "01FN3EEB2NVFJAHAPU00000001",
			reqBody: &models.User{
				Id:       "01FN3EEB2NVFJAHAPU00000001",
				Name:     PointerString("michael"),
				Mail:     "firstuser@mail.com",
				Password: "MyPassword.123",
			},
			expectedResp: &models.User{
				Id:       "01FN3EEB2NVFJAHAPU00000001",
				Name:     PointerString("michael"),
				Mail:     "firstuser@mail.com",
				Password: "",
			},
			expectedStatusCode: http.StatusOK,
			wantErr:            false,
		},
		{
			name:   "[002] Update user that does not exists (404)",
			userID: "01FN3EEB2NVFJAHAPU00000099",
			reqBody: &models.User{
				Name:     PointerString("invent"),
				Mail:     "inventuser@mail.com",
				Password: "MyPassword.123",
			},
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusNotFound,
					Message: internal.ErrUserNotFound.Error(),
				},
			},
			expectedStatusCode: http.StatusNotFound,
			wantErr:            true,
		},
		{
			name: "[003] User id not indicated (400)",
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusBadRequest,
					Message: internal.ErrUserIDNotPresent.Error(),
				},
			},
			expectedStatusCode: http.StatusBadRequest,
			wantErr:            true,
		},
		{
			name:    "[004] Wrong struct sent (400)",
			userID:  "01FN3EEB2NVFJAHAPU00000001",
			reqBody: "invalid",
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusBadRequest,
					Message: internal.ErrWrongBody.Error(),
				},
			},
			expectedStatusCode: http.StatusBadRequest,
			wantErr:            true,
		},
	}
	getEchoContext := func(userId string, request interface{}) echo.Context {
		var body []byte
		body, err := jsoniter.Marshal(request)
		s.NoError(err)
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, internal.RouteUserID, bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames(internal.ParamUserID)
		c.SetParamValues(userId)
		return c
	}
	for _, t := range tests {
		s.Run(t.name, func() {
			userManager := managers.NewUserManager(*s.db)
			api := UserAPI{DB: *s.db, Manager: userManager}

			c := getEchoContext(t.userID, t.reqBody)
			err := api.PutUserHandler(c)

			if t.wantErr {
				s.Equal(t.wantErr, err != nil)
				resp, ok := c.Response().Writer.(*httptest.ResponseRecorder)
				s.True(ok)
				body := resp.Body.Bytes()

				errorReturned := new(internal.ErrorResponse)
				s.NoError(jsoniter.Unmarshal(body, errorReturned))
				s.Equal(errorReturned, t.expectedResp)
			} else {
				resp, ok := c.Response().Writer.(*httptest.ResponseRecorder)
				s.True(ok)
				body := resp.Body.Bytes()

				actualUser := new(models.User)
				s.NoError(jsoniter.Unmarshal(body, actualUser))
				s.Equal(actualUser, t.expectedResp)
			}

			s.Equal(t.expectedStatusCode, c.Response().Status)
		})
	}
}

func (s *UserAPITestSuite) TestDeleteUserHandler() {
	tests := []struct {
		name               string
		userId             string
		expectedResp       interface{}
		expectedStatusCode int
		wantErr            bool
	}{
		{
			name:               "[001] Delete user (ok)",
			userId:             "01FN3EEB2NVFJAHAPU00000001",
			expectedStatusCode: http.StatusNoContent,
			wantErr:            false,
		},
		{
			name:   "[002] Delete user that does not exists (404)",
			userId: "01FN3EEB2NVFJAHAPU00000099",
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusNotFound,
					Message: internal.ErrUserNotFound.Error(),
				},
			},
			expectedStatusCode: http.StatusNotFound,
			wantErr:            true,
		},
		{
			name: "[003] User id not indicated (400)",
			expectedResp: &internal.ErrorResponse{
				Err: internal.ErrorBody{
					Status:  http.StatusBadRequest,
					Message: internal.ErrUserIDNotPresent.Error(),
				},
			},
			expectedStatusCode: http.StatusBadRequest,
			wantErr:            true,
		},
	}
	getEchoContext := func(userId string) echo.Context {
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, internal.RouteUserID, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames(internal.ParamUserID)
		c.SetParamValues(userId)
		return c
	}
	for _, t := range tests {
		s.Run(t.name, func() {
			userManager := managers.NewUserManager(*s.db)
			api := UserAPI{DB: *s.db, Manager: userManager}

			c := getEchoContext(t.userId)
			err := api.DeleteUserHandler(c)

			if t.wantErr {
				s.Equal(t.wantErr, err != nil)
				resp, ok := c.Response().Writer.(*httptest.ResponseRecorder)
				s.True(ok)
				body := resp.Body.Bytes()

				errorReturned := new(internal.ErrorResponse)
				s.NoError(jsoniter.Unmarshal(body, errorReturned))
				s.Equal(errorReturned, t.expectedResp)
			}
			s.Equal(t.expectedStatusCode, c.Response().Status)
		})
	}
}

func PointerString(v string) *string { return &v }
