package empcron

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestNewTimeMask(t *testing.T) {
	tm, _ := NewTimeMask(map[string][]int{"Min": []int{5, 10, 15}})
	tYes, _ := time.Parse("2006-01-02 15:04:05", "2017-10-23 16:10:00")
	tNo, _ := time.Parse("2006-01-02 15:04:05", "2017-10-23 16:11:00")
	if !tm.isNow(tYes) {
		t.Error("Error #1.1")
	}
	if tm.isNow(tNo) {
		t.Error("Error #1.2")
	}

	tm, _ = NewTimeMask(map[string][]int{"Min": []int{5, 10, 15}, "Hour": []int{10, 12, 14, 16}})
	tYes, _ = time.Parse("2006-01-02 15:04:05", "2017-10-23 16:10:00")
	tNo, _ = time.Parse("2006-01-02 15:04:05", "2017-10-23 17:10:00")
	if !tm.isNow(tYes) {
		t.Error("Error #2.1")
	}
	if tm.isNow(tNo) {
		t.Error("Error #2.2")
	}

}

func TestNewTimeMaskFromString(t *testing.T) {
	tm, _ := NewTimeMaskFromString("5,10,15 */2 * *")
	// fmt.Printf("%#v\n", tm)
	tYes, _ := time.Parse("2006-01-02 15:04:05", "2017-10-23 16:10:00")
	tNo, _ := time.Parse("2006-01-02 15:04:05", "2017-10-23 16:11:00")
	if !tm.isNow(tYes) {
		t.Error("Error #1.1")
	}
	if tm.isNow(tNo) {
		t.Error("Error #1.2")
	}

	tm, _ = NewTimeMaskFromString("5,10,15 15-18 * *")
	// fmt.Printf("%#v\n", tm)
	tYes, _ = time.Parse("2006-01-02 15:04:05", "2017-10-23 16:10:00")
	tNo, _ = time.Parse("2006-01-02 15:04:05", "2017-10-23 17:11:00")
	if !tm.isNow(tYes) {
		t.Error("Error #2.1")
	}
	if tm.isNow(tNo) {
		t.Error("Error #2.2")
	}

	tm, _ = NewTimeMaskFromString("5-15 * * *")
	// fmt.Printf("%#v\n", tm)
	tYes, _ = time.Parse("2006-01-02 15:04:05", "2017-10-23 16:10:00")
	tNo, _ = time.Parse("2006-01-02 15:04:05", "2017-10-23 17:20:00")
	if !tm.isNow(tYes) {
		t.Error("Error #3.1")
	}
	if tm.isNow(tNo) {
		t.Error("Error #3.2")
	}

	tm, _ = NewTimeMaskFromString("*/3 * * *")
	// fmt.Printf("%#v\n", tm)
	tYes, _ = time.Parse("2006-01-02 15:04:05", "2017-10-23 16:9:00")
	tNo, _ = time.Parse("2006-01-02 15:04:05", "2017-10-23 17:10:00")
	if !tm.isNow(tYes) {
		t.Error("Error #4.1")
	}
	if tm.isNow(tNo) {
		t.Error("Error #4.2")
	}

}

func foo(ctx context.Context) error {
	fmt.Printf("Call foo\n")
	var (
		s string
	)
	db := ctx.Value("DB").(*sql.DB)
	db.QueryRow("select description from category where id=4").Scan(&s)
	fmt.Printf("\t\t\tfoo=%s!\n", s)
	fmt.Printf("\t\t\tctx.testParam=%d!\n", ctx.Value("testParam"))
	fmt.Printf("\t\t\tctx.testParam2=%d!\n", ctx.Value("testParam2"))
	return nil
}

func slowFunc(ctx context.Context) error {
	n := time.Now().Format("04")
	fmt.Println("Call slowFunc:", n)
	defer fmt.Println("Finish slowFunc:", n)
	fmt.Println(ctx.Value("testParam"))
	time.Sleep(65 * time.Second)
	fmt.Println(ctx.Value("anotherTestParam"))
	return nil
}

func fire(ctx context.Context) error {
	fmt.Println("Call fire")
	fmt.Println(ctx.Value("testParam"))
	fmt.Println(ctx.Value("anotherTestParam"))
	return nil
}

func TestAddTask(t *testing.T) {
	tm, _ := NewTimeMask(map[string][]int{"Min": []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58}})
	c := NewCron(b())
	id, _ := c.AddTask(tm, foo)

	c.AddParam(id, "testParam", 222)
	c.AddParam(id, "testParam2", 2222)

	fmt.Printf("%#v\n", c)
	c.StartCron()
	time.Sleep(time.Minute * 3)
	c.DownCron()
}

func TestQuitTask(t *testing.T) {
	// tm, _ := NewTimeMask(map[string][]int{"Min": []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58}})
	tm, _ := NewTimeMaskFromString("* * * *")
	c := NewCron(b())
	id, _ := c.AddTask(tm, foo)

	c.AddParam(id, "testParam", 22)
	c.AddParam(id, "testParam2", 2)

	fmt.Printf("%#v\n", c)
	c.StartCron()
	time.Sleep(time.Minute * 3)
	c.Quit <- true
}

func TestFire(t *testing.T) {
	tm, _ := NewTimeMask(map[string][]int{"Min": []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58}})
	c := NewCron(b())
	id, _ := c.AddTask(tm, fire)
	c.AddParam(id, "testParam", "it works1")
	c.AddParam(id, "anotherTestParam", "it works2")
	c.StartCron()
	for i := 0; i < 10; i++ {
		c.Fire(id)
	}
	c.DownCron()
}

func TestSlowFunc(t *testing.T) {
	tm, _ := NewTimeMaskFromString("* * * *")
	c := NewCron(b())
	id, _ := c.AddTask(tm, slowFunc)
	c.AddParam(id, "testParam", "Slow1")
	c.AddParam(id, "anotherTestParam", "Slow2")
	c.StartCron()
	time.Sleep(time.Minute * 10)
	c.DownCron()
}

func b() *sql.DB {
	db, _ := sql.Open("postgres", "postgres://postgres@10.250.29.19:5432/emp_admin_app?sslmode=disable")
	return db
}
