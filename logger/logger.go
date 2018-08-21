package alxlogger

import (
	"bufio"
	"fmt"
	// "parti/rmq"
	"io"
	"log/syslog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LogFunc - тип фунция для вывода логов
type LogFunc func(i syslog.Priority, s string, wg *sync.WaitGroup)

const (
	LOG_ERR                = syslog.LOG_ERR
	LOG_WARNING            = syslog.LOG_WARNING
	LOG_NOTICE             = syslog.LOG_NOTICE
	LOG_INFO               = syslog.LOG_INFO
	LOG_DEBUG              = syslog.LOG_DEBUG
	SYSLOGMAXSTRGINGLENGTH = 512
)

var (
	syslogger       *syslog.Writer
	prefix, binName string
	level           = LOG_DEBUG
	loggers         map[string]LogFunc
	debug           = false
	loggerVersion   = 1
	logChannel      = make(chan string)
	ErrChannel      = make(chan string)
	WriteLogFile    = make(chan bool)
	WriteErrFile    = make(chan bool)
	OutFName        string
	OutFDesc        *os.File
	ErrFileDesc     error
	ErrFDesc        *os.File
	ErrFName        string
	FileError       error
	FileErr2        *os.File
	// rmqLogCli       *rmq.Cli
	mutex sync.Mutex
)

func init() {
	loggers = make(map[string]LogFunc)
	if binName == "" {
		p := strings.Split(os.Args[0], "/")
		binName = p[len(p)-1]
	}
	EnableSyslog()
	AddLogStream(os.Stderr)
}

// SetVersion смена режима работы модуля
func SetVersion(v int) {
	mutex.Lock()
	loggerVersion = v
	mutex.Unlock()
}

// Close Завершение работы логгера, defer it
func Close() {
	DisableSyslog()
	DisableLogStream("stderr")
}

// EnableSyslog Открываем коннект к сислогу
func EnableSyslog() *syslog.Writer {
	var (
		err error
	)

	syslogger, err = syslog.New(syslog.LOG_DEBUG, binName)
	if err != nil {
		panic(err)
	}

	AddLogger("syslog", sysLogFunc)
	return syslogger
}

// getErrorTexts - получение текста ошибки
func getErrorTexts(key int) (string, string, int) {
	if e, ok := ErrorList[key]; ok {
		if e.PublicText == "" {
			e.PublicText = ErrorList[key/1000*1000].PublicText
		}
		if e.PublicErrCode == 0 {
			e.PublicErrCode = ErrorList[key/1000*1000].PublicErrCode
		}
		return e.InnerText, e.PublicText, e.PublicErrCode
	}
	// return NotInListErrorInner, NotInListErrorPublic, NotInListErrorCode
	return "", "", 0
}

// RemoveLogger удаление логера из списка
func RemoveLogger(key string) {
	mutex.Lock()
	delete(loggers, key)
	mutex.Unlock()
}

// AddLogger добавление логера в список
func AddLogger(key string, f LogFunc) {
	mutex.Lock()
	loggers[key] = f
	mutex.Unlock()
}

// DisableSyslog Закрываем коннект к сислогу
// при закрытии приложения явно вызывать необязательно - функция встроена в Close()
func DisableSyslog() {
	syslogger.Close()
	RemoveLogger("syslog")
}

// AddLogStream Включаем логирование в передаваемый writer
// параметр оставлен для совместимости
func AddLogStream(stream io.Writer) {
	AddLogger("stderr", stderrLogFunc)
}

// DisableLogStream Отключаем логирование в переданный writer (входит в Close())
func DisableLogStream(key string) {
	if _, ok := loggers[key]; ok {
		RemoveLogger(key)
	}
}

// SetLogLevel Устанавливаем уровень логирования
func SetLogLevel(incomingLevel int) {
	mutex.Lock()

	switch incomingLevel {
	case 3:
		level = LOG_ERR
	case 4:
		level = LOG_WARNING
	case 5:
		level = LOG_NOTICE
	case 6:
		level = LOG_INFO
	case 7:
		level = LOG_DEBUG
	default:
		Err(10000, "Unexpected log level")
	}
	mutex.Unlock()
}

// GetTextLogLevel Получаем название уровня логирования
func GetTextLogLevel() string {
	switch level {
	case LOG_ERR:
		return "LOG_ERR"
	case LOG_WARNING:
		return "LOG_WARNING"
	case LOG_INFO:
		return "LOG_INFO"
	case LOG_DEBUG:
		return "LOG_DEBUG"
	case LOG_NOTICE:
		return "LOG_NOTICE"
	}
	return ""
}

// GetLogLevel Получаем уровень логирования
func GetLogLevel() int {
	return int(level)
}

// SetPrefix Устанавливаем префиксErrFileDesc
func SetPrefix(incomingPrefix string) {
	mutex.Lock()
	prefix = incomingPrefix
	mutex.Unlock()
}

// GetPrefix получаем префикс
func GetPrefix() string {
	return prefix
}

// Функция логирования
// Принимает:
//  level - уровень логирования
//  v - variadic parameter с логируемыми данными
// Логирует полученные данные, если переданный уровень логирования равен или больше,
// чем установленный в данный момент в логгере
func log(incomingLevel syslog.Priority, v ...interface{}) {
	if level < incomingLevel {
		return
	}

	var (
		wg sync.WaitGroup
	)

	syslogString := fmt.Sprint(prefix, " ", v)

	for _, f := range loggers {
		wg.Add(1)
		go f(incomingLevel, syslogString, &wg)
	}

	//запись в канал ошибок l.Err, l.Warn, panic
	if ErrFDesc != nil && (incomingLevel == 3 || incomingLevel == 4) {
		wg.Add(1)
		go SendErrToChannel(&wg, syslogString)
	}
	if OutFDesc != nil {
		wg.Add(1)
		go sendLogToChannel(&wg, syslogString)
	}
	wg.Wait()
}

// sysLogFunc LogFunc для локального syslog
func sysLogFunc(incomingLevel syslog.Priority, s string, wg *sync.WaitGroup) {
	defer wg.Done()

	if debug {
		s = "syslog::" + s
	}

	l := len(s)
	if l > SYSLOGMAXSTRGINGLENGTH {
		l = SYSLOGMAXSTRGINGLENGTH
	}
	s = s[:l]
	switch incomingLevel {
	/*case syslog.LOG_CRIT:
	  syslogger.Crit(syslogString)*/
	case LOG_ERR:
		syslogger.Err(s)
	case LOG_WARNING:
		syslogger.Warning(s)
	case LOG_INFO:
		syslogger.Info(s)
	case LOG_DEBUG:
		syslogger.Debug(s)
	default:
		logPanic(LOG_ERR, "You have used forbidden log level")
	}
}

//OutFDesc - файловый дескриптор доступный в main файле
func SetLogFileDesc(logFile string) error {
	mutex.Lock()
	OutFName = logFile
	mutex.Unlock()
	OutFDesc, ErrFileDesc = os.OpenFile(OutFName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	return ErrFileDesc
}

// SetErrFileDesc метод создает файл для ошибок уровня err, warning, panic
func SetErrFileDesc(errorFile string) error {
	mutex.Lock()
	ErrFName = errorFile
	mutex.Unlock()
	ErrFDesc, FileError = os.OpenFile(ErrFName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	return FileError
}

//LogFileInit - инициализация процесса записи в файл
func LogFileInit() {
	go func() {
		for {
			select {
			case lc, ok := <-logChannel:
				if !ok {
					Err("Ошибка чтения log канала")
					break
				}
				mutex.Lock()
				fmt.Fprint(OutFDesc, lc)
				mutex.Unlock()
			case wtf := <-WriteLogFile:
				if !wtf {
					Err("Файл закрыт:" + OutFName)
					OutFDesc.Close()
				}
				if wtf {
					Info("Файл открыт" + OutFName)
					OutFDesc, ErrFileDesc = os.OpenFile(OutFName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
					if ErrFileDesc != nil {
						Err(11005, fmt.Sprintf("файл: %s, ошибка: %s", OutFName, ErrFileDesc.Error()))
					}
				}
			}
		}
	}()
}

//ErrFileInit - инициализация процесса записи в файл l.Err, l.Warn, panic
func ErrFileInit() {
	go func() {
		for {
			select {
			case lc, ok := <-ErrChannel:
				if !ok {
					Err("Ошибка чтения %s канала log")
					break
				}

				mutex.Lock()
				fmt.Fprint(ErrFDesc, lc)
				mutex.Unlock()

			case wtf := <-WriteErrFile:
				if !wtf {
					Err("Файл закрыт:" + ErrFName)
					OutFDesc.Close()
				}
				if wtf {
					Info("Открываем файл: " + ErrFName)
					ErrFDesc, FileError = os.OpenFile(ErrFName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
					if FileError != nil {
						Err(11005, "файл:%s, ошибка:%s", OutFName, ErrFileDesc.Error())
					}
				}
			}
		}
	}()
}

func writeLogIntoFile(s string, f *os.File) {
	fmt.Fprint(f, s)
	// f.Sync()
	// w := bufio.NewWriter(f)
	// w.WriteString(s)
	// w.Flush()
}

func WriteErrIntoFile(s string, f *os.File) {

	f.Sync()
	w := bufio.NewWriter(f)
	w.WriteString(s)
	w.Flush()
}

func sendLogToChannel(wg *sync.WaitGroup, v ...interface{}) {
	defer wg.Done()
	//на вход может прийти error
	strLog := fmt.Sprint(v, "\n")
	logChannel <- fmt.Sprintf("[%s] %s", time.Now().Format("Jan 02 15:04:05.999999999"), strLog)
}

// SendErrToChannel метод отправки ошибки в лог уровня l.Err, l.Warn, panic
func SendErrToChannel(wg *sync.WaitGroup, v ...interface{}) {
	defer wg.Done()
	//на вход может прийти error
	strLog := fmt.Sprint(v, "\n")
	ErrChannel <- fmt.Sprintf("[%s] %s", time.Now().Format("Jan 02 15:04:05.999999999"), strLog)
}

// stderrLogFunc LogFunc для локального stderr
func stderrLogFunc(incomingLevel syslog.Priority, s string, wg *sync.WaitGroup) {
	defer wg.Done()

	if debug {
		s = "stderr::" + s
	}

	fmt.Fprintf(os.Stderr, "%s\n", s)
}

func getErrCode(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	return i
}

// Err набор функций, вызывающих Log() с фиксированным уровнем логирования
func Err(v ...interface{}) (string, int) {
	var (
		privateText, publicText string
		publicErrCode           int
	)

	if loggerVersion < 2 {
		log(LOG_ERR, v...)
	} else {
		switch v[0].(type) {
		case int:
			privateText, publicText, publicErrCode = getErrorTexts(v[0].(int))
			log(LOG_ERR, v[0].(int), privateText, v[1:])
		case string:
			k := getErrCode(v[0].(string))
			if k != -1 {
				privateText, publicText, publicErrCode = getErrorTexts(k)
				log(LOG_ERR, v[0].(string), privateText)
			} else {

				log(LOG_ERR, v...)
			}
		}
	}
	return publicText, publicErrCode
}

// Warn функиция предупреждения
func Warn(v ...interface{}) (string, int) {
	var (
		privateText, publicText string
		publicErrCode           int
	)

	if loggerVersion < 2 {
		log(LOG_WARNING, v...)
	} else {
		switch v[0].(type) {
		case int:
			privateText, publicText, publicErrCode = getErrorTexts(v[0].(int))
			log(LOG_WARNING, v[0].(int), privateText, v[1:])
		case string:
			k := getErrCode(v[0].(string))
			if k != -1 {
				privateText, publicText, publicErrCode = getErrorTexts(k)
				log(LOG_WARNING, v[0].(string), privateText)
			} else {
				log(LOG_WARNING, v...)
			}
		}
	}
	return publicText, publicErrCode
}

// Info функция вывода информции
func Info(v ...interface{}) (string, int) {
	var (
		privateText, publicText string
		publicErrCode           int
	)

	if loggerVersion < 2 {
		log(LOG_INFO, v...)
	} else {
		switch v[0].(type) {
		case int:
			privateText, publicText, publicErrCode = getErrorTexts(v[0].(int))
			log(LOG_INFO, v[0].(int), privateText, v[1:])
		case string:
			k := getErrCode(v[0].(string))
			if k != -1 {
				privateText, publicText, publicErrCode = getErrorTexts(k)
				log(LOG_INFO, v[0].(string), privateText)
			} else {
				log(LOG_INFO, v...)
			}
		}
	}
	return publicText, publicErrCode
}

// Dump - вывод больших данных
func Dump(title string, m map[string]interface{}) {
	out := "\n" + title + "\n"
	for k, v := range m {
		println(k)
		s := ""
		switch v.(type) {
		case string:
			s = v.(string)
		case []byte:
			s = string(v.([]byte))
		default:
			s = fmt.Sprintf("%#v", v)
		}
		if s != "" {
			out = out + "\t" + k + ":\n\t\t" + s + "\n"
		}
	}

	log(LOG_DEBUG, out)
}

// Debug функция вывода отладочной информации
func Debug(v ...interface{}) (string, int) {
	var (
		privateText, publicText string
		publicErrCode           int
	)

	if loggerVersion < 2 {
		log(LOG_DEBUG, v...)
	} else {
		switch v[0].(type) {
		case int:
			privateText, publicText, publicErrCode = getErrorTexts(v[0].(int))
			log(LOG_DEBUG, v[0].(int), privateText, v[1:])
		case string:
			k := getErrCode(v[0].(string))
			if k != -1 {
				privateText, publicText, publicErrCode = getErrorTexts(k)
				log(LOG_DEBUG, v[0].(string), privateText)
			} else {
				log(LOG_DEBUG, v...)
			}
		}
	}
	return publicText, publicErrCode
}

// Лоигирование с последующим вызовом panic()
func logPanic(level syslog.Priority, v ...interface{}) {
	log(level, v...)
	panic(v)
}

// ErrPanic набор функций, вызывающих LogPanic() с фиксированным уровнем логирования
func ErrPanic(v ...interface{}) {
	logPanic(LOG_ERR, v...)
}

// WarnPanic Warn + panic
func WarnPanic(v ...interface{}) {
	logPanic(LOG_WARNING, v...)
}

// InfoPanic Info + panic
func InfoPanic(v ...interface{}) {
	logPanic(LOG_INFO, v...)
}

// DebugPanic Debug + panic
func DebugPanic(v ...interface{}) {
	logPanic(LOG_DEBUG, v...)
}
