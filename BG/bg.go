package BG

/*
Чтобы активировать крон надо снять комментарии и заменить возврат функции cronInit на true
*/

import (
	"context"
	"errors"
	"parti/APPLIB"
	cron "parti/emp-go-cron"
	// "database/sql"
    "fmt"
)


// CronInit - функция запуска ФП
// ВНИМАНИЕ !! Если вернуть true, то крон запустится, если нет, то нет
func CronInit(ctx context.Context) (context.Context, bool, error) {

	cTmp := ctx.Value("CRON")
	if cTmp == nil {
		return ctx, false, errors.New("Не определен параметр CRON")
	}
	c := cTmp.(*cron.CronType)

	// cfgTmp := ctx.Value("CFG")
	// if cfgTmp == nil {
	// 	return ctx, false, errors.New("Не определен параметр CFG")
	// }
	// cfg := cfgTmp.(config.Config)

	// dbTmp := ctx.Value("DB")
	// if dbTmp == nil {
	// 	return ctx, false, errors.New("Не определен параметр DB")
	// }
	// db := dbTmp.(*sql.DB)
    
    cancelCTXTmp := ctx.Value("CTX")
    cancelCTX := ctx
    if cancelCTXTmp == nil {
        fmt.Errorf("Ошибка: в контексте нет параметра CTX")
    } else {
        cancelCTX = cancelCTXTmp.(context.Context)
    }
    
    // очередь ошибок
    errChan := make(chan string)
    ctx = context.WithValue(ctx, "errorChan", errChan)
    
    // очередь debug
    debugChan := make(chan string)
    ctx = context.WithValue(ctx, "debugChan", debugChan)
    
    // очередь info
    infoChan := make(chan string)
    ctx = context.WithValue(ctx, "infoChan", infoChan)
    
    go loggerChans(errChan, debugChan, infoChan, cancelCTX.Done())
	
	
	// пример крона
	tm, _ := cron.NewTimeMaskFromString("* * * *")
	c.AddTask(tm, func(ctx context.Context) error {
		println("\n\tStart test cron process")
		dbTmp := APPLIB.DataCtx.Value("Name").(string)
		println("\t\tПолучили из APPLIB.DataCtx::", dbTmp, "\n")
		return nil
	})
	// c.Fire(id)

	// ВНИМАНИЕ !! Если вернуть true, то крон запустится, если нет, то нет
	return ctx, true, nil
}
