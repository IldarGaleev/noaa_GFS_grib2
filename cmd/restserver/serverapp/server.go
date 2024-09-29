package serverapp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrHttpAppRunError  = errors.New("http app run error")
	ErrHttpAppNotRun    = errors.New("http app not run")
	ErrHttpAppStopError = errors.New("http app stop error")
)

type WKTHandler interface {
	HandlerByWKT(c *gin.Context)
}

type ServerApp struct {
	srv    *http.Server
	router *gin.Engine
}

func New(
	apiBasePath string,
	wktHandler WKTHandler,
) *ServerApp {

	router := gin.Default()
	apiNoAuth := router.Group(apiBasePath)
	apiNoAuth.POST("/bywkt", wktHandler.HandlerByWKT)

	return &ServerApp{
		router: router,
	}
}

func (app *ServerApp) Run(host string, port int) error {

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: app.router.Handler(),
	}

	app.srv = srv
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Join(ErrHttpAppRunError, err)
	}

	return nil
}

func (app *ServerApp) Stop(ctx context.Context) error {
	if app.srv == nil {
		return ErrHttpAppNotRun
	}

	err := app.srv.Shutdown(ctx)
	if err != nil {
		return errors.Join(ErrHttpAppStopError, err)
	}

	return nil
}
