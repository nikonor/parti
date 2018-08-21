package bar

import (
    "database/sql"
    "parti/api"
    l "parti/logger"
    "time"
    
    "github.com/labstack/echo"
)


func SleepingFoo(c echo.Context) error {
	b := time.Now().Format("15:04:05")
	time.Sleep(5 * time.Second)
	b = b + " -  " + time.Now().Format("15:04:05")
	api.Response(c, b)
	return nil
}

// Foo пример функции обработчика вызова.
func Foo(c echo.Context) error {
	l.Debug("Вызов bar.Foo")
	defer l.Debug("Конец bar.Foo")
	defer c.Request().Body.Close()
	var (
		err  error
		in   Type
		out  OutType
		port string
		tx   *sql.Tx
	)

	if ec, em := api.UniversalBegin(c, &in); ec != 0 {
		api.Fault(c, ec, em)
		return nil
	}

	// получаем соедиение из пула
	pDB := c.Get("PoolDB").(*api.DBPool)
	// cfg := c.Get("CFG").(*config.Config)
	db, dbID := pDB.GetConn()
	defer pDB.GiveBack(dbID)

	if tx, err = db.Begin(); err != nil {
		api.Fault(c, 17005, err.Error())
		return nil
	}
	defer tx.Rollback()

	// if port, err = foo(db); err != nil {
	if port, err = foo(tx); err != nil {
		api.Fault(c, 17004, err.Error())
		return nil
	}

	out = OutType{Groups: in.Groups, Port: port, ID: in.ID, Name: in.Name}
	api.Response(c, out)
	return nil
}

func foo(aDB api.AccessToDB) (string, error) {
	var (
		ret string
		err error
	)

	if err = aDB.QueryRow("select data->'module'->'http'->'port' from config where is_active='t'").Scan(&ret); err != nil {
		return "", err
	}
	return ret, nil
}

// Bar - возвращает ошибку для FaultF
func Bar(c echo.Context) error {
	l.Debug("Вызов Bar")
	defer l.Debug("Конец Bar")
	defer c.Request().Body.Close()

	api.FaultF(c, 10913, "QQQ", []string{"22", "ку-ку-ку"})

	return nil
}
