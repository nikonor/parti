package main

import "github.com/labstack/echo"

import (
	"context"
	"errors"
	"parti/api"
	"parti/config"
)

func echoInit(dCtx context.Context) (context.Context, error) {
	eTmp := dCtx.Value("ECHO")
	if eTmp == nil {
		return dCtx, errors.New("Не определен параметр ECHO")
	}
	e := eTmp.(*echo.Echo)

	cfgTmp := dCtx.Value("CFG")
	if cfgTmp == nil {
		return dCtx, errors.New("Не определен параметр CFG")
	}
	cfg := cfgTmp.(config.Config)

	dbPoolTmp := dCtx.Value("DBPOOL")
	if dbPoolTmp == nil {
		return dCtx, errors.New("Не определен параметр DBPOOL")
	}
    dbPool := dbPoolTmp.(api.DBPool)
    
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
            
            c.Set("PoolDB", &dbPool)
            c.Set("CFG", &cfg)
			if err := next(c); err != nil {
				return err
			}
			return nil
		}
	})

	return dCtx, nil
}

