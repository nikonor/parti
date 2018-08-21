package sigchan

import (
	"os"
	"syscall"
)

var (
	SigChan chan os.Signal
)

func init() {
	SigChan = make(chan os.Signal, 1)
}

// Send - отправка произвольного сигнала
func Send(s syscall.Signal) {
	SigChan <- s
}

// Reload - отправка команды на перезагрузку модуля
func Reload() {
	SigChan <- syscall.SIGUSR1
}

// ShutDown - отправка команды на выключение модуля
func ShutDown() {
	SigChan <- syscall.SIGTERM
}

// // StopBG - отправка команды на выключение фоновых процессов
// func StopBG() {
// 	SigChan <- syscall.SIGURG
// }

// // StartBG - отправка команды на включение фоновых процессов
// func StartBG() {
// 	SigChan <- syscall.SIGALRM
// }
