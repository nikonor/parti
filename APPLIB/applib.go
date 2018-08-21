package APPLIB

import (
	"context"
	"errors"
	"parti/api"
	"parti/bar"
	"time"

	"github.com/labstack/echo"
)

// "parti/config"

var (
	DataCtx context.Context
)

// WaitHandlerIn - ...
type WaitHandlerIn struct {
	Seconds int `json:"seconds"`
}

// WaitHandlerOut - ...
type WaitHandlerOut struct {
	Seconds int `json:"seconds"`
}

// ApplyRoutes ...
func ApplyRoutes(ctx context.Context) (context.Context, error) {
	// cfgTmp := ctx.Value("CFG")
	// if cfgTmp == nil {
	// 	return ctx, errors.New("Не определена CFG")
	// }
	// cfg := cfgTmp.(config.Config)

	eTmp := ctx.Value("ECHO")
	if eTmp == nil {
		return ctx, errors.New("Не определена ECHO")
	}
	e := eTmp.(*echo.Echo)

	rTmp := ctx.Value("R")
	if rTmp == nil {
		return ctx, errors.New("Не определена R")
	}
	r := rTmp.(*api.RouterHandlers)

	e.GET("/ok", api.OkHandler)
	api.POST(e, r, "/wait", WaitHandler, []interface{}{WaitHandlerIn{}}, []interface{}{WaitHandlerOut{}}, "WaitHandler", "")
	
	api.POST(e, r, "/foo",
		bar.Foo,
		[]interface{}{bar.Type{}, bar.GroupType{}},
		[]interface{}{bar.OutType{}, bar.GroupType{}},
		"Foo", "")

	api.GET(e, r, "/foo/bar", bar.Bar, nil, nil, "", "")
	api.GET(e, r, "/foo/sleep", bar.SleepingFoo, nil, nil, "", "")

	e.GET("/fail", FailHandler)
	e.GET("/pg/read", PgReadHandler)
	e.POST("/pg/write", PgWriteHandler)
	e.POST("/mq/pub", RmqPubHandler)

	return ctx, nil
}

// WaitHandler - тестовый обработчик, успешный ответ с ожиданием
func WaitHandler(c echo.Context) error {
	// @TODO: валидировать входные данные

	data := WaitHandlerOut{Seconds: 5}

	time.Sleep(5 * time.Second)
	return api.Response(c, data)
}

// FailHandler - тестовый обработчик, ответ с ошибкой
func FailHandler(c echo.Context) error {
	h := make(map[string]string)
	h["X-APP-SOME-KEY-1"] = "value-for-key-1"
	h["X-APP-KEY-2"] = "value:for:key:2"
	return api.Fault(c, 888, "Выдуманная ошибка", h)
}

// PgWriteHandler .....
func PgWriteHandler(c echo.Context) error {
	return api.Fault(c, 999, "Не реализовано")
}

// PgReadHandler .....
func PgReadHandler(c echo.Context) error {
	return api.Fault(c, 999, "Не реализовано")
}

// RmqPubHandler .....
func RmqPubHandler(c echo.Context) error {
	return api.Fault(c, 999, "Не реализовано")
}

// TheEnd - финальная функция
func TheEnd(data context.Context) {
	println("Call TheEnd")
}
