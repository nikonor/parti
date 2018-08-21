package api

import (
	"database/sql"
	"sync"
)

type dbLocker struct {
	DB *sql.DB
	sync.Mutex
}

// DBPool - тип для хранения пула соединений с БД
type DBPool struct {
	DBs        []dbLocker
	ArendaList map[*sql.DB]int
	Count      int
	Last       int
	WillLock   bool
	sync.Mutex
}

// Close - закрываем соединения из пула БД
func (p *DBPool) Close() error{
    var (
        err error
    )
    p.Lock()
    defer p.Unlock()
    
    for _,d := range p.DBs{
        d.Lock()
        if err = d.DB.Close();  err != nil {
            return err
        }
        d.Unlock()
    }
    return nil
}


// GetConn - получение соединения с БД из пула
func (p *DBPool) GetConn() (*sql.DB, int) {
	l := p.Last + 1
	if l >= p.Count {
		l = 0
	}

	p.Lock()
	p.Last = l
	// p.ArendaList[p.DBs[l]] = l
	if p.WillLock {
		p.DBs[l].Lock()
	}
	p.Unlock()

	p.DBs[l].DB.Ping()

	return p.DBs[l].DB, l
}

// AddConn - добавление соединения в пул БД
func (p *DBPool) AddConn(db *sql.DB) {
	p.Lock()
	defer p.Unlock()
	p.DBs = append(p.DBs, dbLocker{DB: db})
	p.Count++
}

// NewPool создание нового пула
//	если параметр передан и он true, то выданное соединение будет лочится
//	в этом случае ОБЯЗАТЕЛЬНО использовать GiveBack
func NewPool(willLock ...bool) *DBPool {
	r := DBPool{}

	if len(willLock) == 1 {
		r.WillLock = willLock[0]
	} else {
		r.WillLock = false
	}
	// r.ArendaList = make(map[*sql.DB]int)
	return &r
}

// GiveBack - возвращаем соединение в пул БД
func (p *DBPool) GiveBack(l int) {
	if p.WillLock {
		p.DBs[l].Unlock()
	}
}

// FillPool - заполняем пулл соединенями
func (p *DBPool) FillPool(c int, dsn string) error {
    var (
        db *sql.DB
        err error
    )
    for i := 0; i < c; i++ {
        // Подключаемся к БД
        db, err = sql.Open("postgres", dsn)
        if err != nil {
            return err
        }
        
        if err = db.Ping(); err != nil {
            return err
        }
        
        p.AddConn(db)
    }
    return nil
}