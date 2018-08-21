package alxlogger

import (
	// "fmt"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	Err("Первое сообщение")
	SetPrefix("logger-test::")
	Err("Второе сообщение")
	Debug("Третье сообщение")
	SetPrefix("PrEfIx")
	Warn("Четвертое сообщение")
}

func TestLevel(t *testing.T) {
	Debug("Пятое сообщение")
	SetLogLevel(3)
	// Шестое сообщение не появляется нигде
	Debug("Шестое сообщение")
	SetLogLevel(7)
}

func TestGetErrorTexts(t *testing.T) {
	t1, t2, t3 := getErrorTexts(11001)
	if t1 != t2 || t1 != "Lock file error" || t3 != 500 {
		t.Errorf("Error #1: %s != %s or %d != 500", t1, t2, t3)
	}

	t1, t2, t3 = getErrorTexts(11002)
	if t1 == t2 || t1 != "Config file not found" || t2 != "Ошибка дисковой подсистемы" || t3 != 404 {
		t.Errorf("Error #2: %s == %s or %d != 404 ", t1, t2, t3)
	}
}

func TestVer2(t *testing.T) {
	s4, _ := Err(11001, "10ое сообщение")
	s5, _ := Err(11002, "11ое сообщение")
	s6, _ := Err(11003, "12ое сообщение")

	if s4 != s5 || s5 != s6 {
		t.Errorf("Error #1: s4=!%s!, s5=!%s!, s6=!%s!", s4, s5, s6)
	}

	SetVersion(2)
	s1, _ := Err(11001, "Седьмое сообщение")
	s2, _ := Err(11002, "Восьмое сообщение")
	s3, _ := Err(11003, "Девятое сообщение")

	if s1 == s2 || s2 != s3 {
		t.Error("Error #2")
	}
}

//func Bench
func BenchmarkLogDebug(b *testing.B) {

	// установка временного префикса для системного журнала
	defer Close()
	SetVersion(2)
	SetLogLevel(7)

	var (
		id      int
		err     error
		errFile = "/tmp/bench.err"
		outFile = "/tmp/bench.out"
	)

	if err = SetErrFileDesc(errFile); err != nil {
		b.Error(err.Error())
	}
	ErrFileInit()

	if err = SetLogFileDesc(outFile); err != nil {
		b.Error(err.Error())
	}
	LogFileInit()

	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		id++
		Debug("Запуск процесса с id = ", id)
		Debug("Завершение процесса с id = ", id)
		Err("Запуск процесса с id = ", id)
		Err("Завершение процесса с id = ", id)
	}

	if err = os.Remove(errFile); err != nil {
		b.Error(err.Error())
	}

	if err = os.Remove(outFile); err != nil {
		b.Error(err.Error())
	}

}
