package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"net/http"
	"users/internal"
	"users/internal/config"
	"users/internal/handlers"
	"users/internal/managers"
	"users/pkg/database"
)

const (
	banner = `
   ___    __  ___  _____       __  __                     
  / _ |  /  |/  / / ___/      / / / /  ___ ___   ____  ___
 / __ | / /|_/ / / /__       / /_/ /  (_-</ -_) / __/ (_-<
/_/ |_|/_/  /_/  \___/       \____/  /___/\__/ /_/   /___/

AMC Users Service
`
)

func main() {

	if err := config.LoadConfiguration(); err != nil {
		log.Fatal(err)
	}
	db := database.InitDB(config.Config.DBName)
	e := setUpServer(db)
	e.Logger.Fatal(e.Start(config.Config.Host + ":" + config.Config.Port))

}

func setUpServer(db *database.Database) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	addRoutes(e, *db)
	e.HideBanner = true
	fmt.Printf(banner)

	return e

}

func addRoutes(e *echo.Echo, db database.Database) {

	userManager := managers.NewUserManager(db)

	userAPI := handlers.UserAPI{DB: db, Manager: userManager}
	e.POST(internal.RouteLogin, userAPI.Login)
	e.POST(internal.RouteUser, userAPI.PostUserHandler)
	e.GET(internal.RouteUserID, userAPI.GetUserHandler)
	e.PUT(internal.RouteUserID, userAPI.PutUserHandler)
	e.DELETE(internal.RouteUserID, userAPI.DeleteUserHandler)
}
