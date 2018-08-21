package empcron

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

//TaskFunc - тип функция для работы
type TaskFunc func(ctx context.Context) error

//TaskTimeMaskType описание времение запуска запланированных процессов
type TaskTimeMaskType struct {
	Min     []int
	Hour    []int
	Day     []int
	Month   []int
	Weekday []int
}

//TaskType тип записи о необходимой работе
type TaskType struct {
	ID       int
	isActive bool
	F        TaskFunc
	T        TaskTimeMaskType
	P        context.Context
}

// CronType структура крона
type CronType struct {
	db     *sql.DB
	ticker *time.Ticker
	tasks  []TaskType
	count  int
	Quit   chan bool
	debug  bool
	active bool
}

// IsCronWork - запущен ли ФП крона
func (c CronType) IsCronWork() bool {
	return c.active
}

// NewCron - получение нового объекта крона
func NewCron(db *sql.DB) *CronType {
	c := CronType{debug: false, db: db}
	c.ticker = time.NewTicker(time.Minute)
	return &c
}

// StartCron - стартуем фоновый процесс
func (c *CronType) StartCron() {
	if !c.active {
		c.active = true
		go func() {
			for {
				select {
				case <-c.Quit:
					c.d("quit")
					c.ticker.Stop()
					return
				case <-c.ticker.C:
					c.d("ticker")
					c.work()
				}
			}
		}()
	}
}

// List - список ФП
func (c CronType) List() []TaskType {
	return c.tasks
}

// AddParam - добавление параметра для процесса
func (c *CronType) AddParam(wID int, key string, val interface{}) {
	c.d("Call AddParam: wID=", wID)
	for i, t := range c.tasks {
		if t.ID == wID {
			if t.P == nil {
				t.P = context.Background()
			}
			c.tasks[i].P = context.WithValue(t.P, key, val)
		}
	}
}

// DownCron - завершение работы
func (c *CronType) DownCron() error {
	c.d("Call DownCron")
	c.ticker.Stop()
	return nil
}

// AddTask - функция добавиления процесса в план
//	возварщает ID процесса
func (c *CronType) AddTask(tm TaskTimeMaskType, f TaskFunc) (int, error) {
	c.d("Call AddTask")
	t := TaskType{isActive: true, ID: c.count + 1, T: tm, F: f, P: context.Background()}
	c.tasks = append(c.tasks, t)
	c.Quit = make(chan bool)
	c.count++
	return c.count, nil
}

// DeActivateTask - функция отключения процесса
func (c *CronType) DeActivateTask(id int) error {
	c.tasks[id-1].isActive = false
	return nil
}

// ReActivateTask - функция включение процесса
func (c *CronType) ReActivateTask(id int) error {
	c.tasks[id-1].isActive = true
	return nil
}

// Fire немедленное выполнение процесса
func (c *CronType) Fire(id int) error {
	task := c.tasks[id-1]
	task.P = context.WithValue(task.P, "DB", c.db)
	return task.F(task.P)
}

// work - основная функция, которую и вызывает тикер
func (c *CronType) work(params ...interface{}) {
	c.d("Call work:", time.Now().Format("15:04:05"))
	for i, t := range c.tasks {
		c.d("\ttask #", i)
		if t.isActive && t.T.isNow(time.Now()) {
			c.d("\t\tRun #", i, t.ID)
			t.P = context.WithValue(t.P, "DB", c.db)
			go func(t TaskType) {
				c.DeActivateTask(t.ID)
				defer c.ReActivateTask(t.ID)
				t.F(t.P)
			}(t)
		} else {
			c.d("\t\tNot run #", i)
		}
	}
}

// NewTimeMaskFromString создание условий по времени для запуска процесса в нотации cron
func NewTimeMaskFromString(in string) (TaskTimeMaskType, error) {
	var (
		tm TaskTimeMaskType
		ss []string
	)

	ss = strings.Fields(in)
	if ss[0] != "*" {
		tm.Min = strToTM(ss[0])
	}
	if ss[1] != "*" {
		tm.Hour = strToTM(ss[1])
	}
	if ss[2] != "*" {
		tm.Day = strToTM(ss[2])
	}
	if ss[3] != "*" {
		tm.Month = strToTM(ss[3])
	}
	if len(ss) > 4 {
		tm.Weekday = strToTM(ss[4])
	}

	return tm, nil
}

func strToTM(s string) []int {
	ret := []int{}
	for _, i := range strings.Split(s, ",") {
		if ii, err := strconv.Atoi(i); err == nil {
			// случай просто перечисления чисел
			ret = append(ret, ii)
		} else {
			if strings.Contains(i, "-") {
				// случай интервала, т.е. 2-12
				sub := strings.SplitN(i, "-", 2)
				subB, errB := strconv.Atoi(sub[0])
				subE, errE := strconv.Atoi(sub[1])
				if errB == nil && errE == nil {
					for j := subB; j <= subE; j++ {
						ret = append(ret, j)
					}
				}
			} else if strings.HasPrefix(i, "*/") {
				// случай маски */2
				if sub, err := strconv.Atoi(strings.SplitN(i, "/", 2)[1]); err == nil {
					for j := 0; j <= 60; j++ {
						if j%sub == 0 {
							ret = append(ret, j)
						}
					}

				}
			}
		}
	}

	return ret
}

// NewTimeMask создание условий по времени для запуска процесса
//	возможные ключи: Min, Hour, Day, Month, Weekday
func NewTimeMask(in map[string][]int) (TaskTimeMaskType, error) {
	var (
		tm TaskTimeMaskType
	)

	for k, v := range in {
		switch k {
		case "Min":
			tm.Min = append(tm.Min, v...)
		case "Hour":
			tm.Hour = append(tm.Hour, v...)
		case "Day":
			tm.Day = append(tm.Day, v...)
		case "Month":
			tm.Month = append(tm.Month, v...)
		case "Weekday":
			tm.Weekday = append(tm.Weekday, v...)
		default:
			return tm, errors.New("Неизвестный ключ")
		}

	}

	return tm, nil
}

// isNow - фунция проверки времени и маски
func (tm TaskTimeMaskType) isNow(t time.Time) bool {
	var (
		yes = 0
	)
	if len(tm.Min) > 0 {
		yes++
		mm := t.Minute()
		for _, m := range tm.Min {
			if m == mm {
				yes--
				break
			}
		}
		if yes > 0 {
			return false
		}
	}
	if len(tm.Hour) > 0 {
		yes++
		mh := t.Hour()
		for _, h := range tm.Hour {
			if h == mh {
				yes--
				break
			}
		}
		if yes > 0 {
			return false
		}
	}
	if len(tm.Day) > 0 {
		yes++
		md := t.Day()
		for _, d := range tm.Day {
			if d == md {
				yes--
				break
			}
		}
		if yes > 0 {
			return false
		}
	}
	if len(tm.Month) > 0 {
		yes++
		mm := t.Month()
		for _, m := range tm.Month {
			if m == int(mm) {
				yes--
				break
			}
		}
		if yes > 0 {
			return false
		}
	}

	if len(tm.Weekday) > 0 {
		yes++
		ww := t.Weekday()
		for _, w := range tm.Weekday {
			if w == int(ww) {
				yes--
				break
			}
		}
		if yes > 0 {
			return false
		}
	}

	return true
}

func (c CronType) d(s ...interface{}) {
	if c.debug {
		fmt.Println(s)
	}
}
