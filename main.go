package main

import (
	"context"
	"flag"
	"fmt"
	"gt2/instance"
	"parti/APPLIB"
	"parti/BG"
	"parti/api"
	"parti/common"
	"parti/config"

	"github.com/fredli74/lockfile"
	_ "github.com/lib/pq"

	_ "github.com/lib/pq"

	cron "parti/emp-go-cron"

	"net"
	"os"
	"os/signal"
	l "parti/logger"
	"parti/sigchan"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo"
)

var (
	moduleVersion     = "1.9"
	moduleName        = "parti"
	dbPool            api.DBPool
	moduleCompileTime = ""
	goTemplateVersion = ""
	gitCommit         = ""
	goBuildVersion    = ""
	goBuildTemplate   = ""
	goVersion         = ""
)

const (
	errServClosed = "http: Server closed"
	MailQueue     = "mail"
)

func main() {
	var (
		cfg            config.Config
		wg             sync.WaitGroup
		err            error
		lock           *lockfile.LockFile
		reload         = true
		netport        = ""
		cpucores       = 0
		originLogLevel int
		updCfgTime     time.Time
	)
	api.StartTime = time.Now()

	// Обработка флагов командной строки
	cfgFlag := flag.String("c", "", "путь к файлу конфигурации")
	flag.StringVar(cfgFlag, "config", "", "путь к файлу конфигурации")
	pidFlag := flag.String("p", fmt.Sprintf("/var/run/emp/%s.pid", moduleName), "путь к PID файлу")
	flag.StringVar(pidFlag, "pid", fmt.Sprintf("/var/run/emp/%s.pid", moduleName), "путь к PID файлу")
	testFlag := flag.Bool("t", false, "валидации конфигурации")
	jsonFlag := flag.Bool("j", false, "вывод конфигурации в формате json")
	dsn := flag.String("pgdsn", "", "строка подключения к PostgreSQL")
	journalName := flag.String("logtag", "", "тэг для системного журнала")
	loglevel := flag.Int("log_level", -1, "уровень логирования. (3-err,4-warn,5-notice,6-info,7-debug")
	appPort := flag.Int64("port", -1, "Сетевой порт WebServer")
	dbPoolSizeFlag := flag.Int("dbpoolsize", -1, "Кол-во соединений с БД в пуле")
	appSocket := flag.String("socket", "", "Unixsocket для WebServer")
	syslogFlag := flag.String("syslog", "on", "вывод в syslog (on-включен, off-выключен")
	stdoutFlag := flag.String("stdout", "on", "вывод в stdout (on-включен, off-выключен")
	pgUser := flag.String("pguser", "", "Пользователь postgres")
	pgPassword := flag.String("pgpassword", "", "Пароль пользователь postgres")
	pgHost := flag.String("pghost", "", "Хост postgres")
	pgPort := flag.String("pgport", "", "Порт postgres")
	pgDataBase := flag.String("pgdatabase", "", "Название БД postgres")
	pgFlags := flag.String("pgflags", "", "Флаги postgres")
	// serviceFlag := flag.Bool("service", false, "Включение сервисного режима")
	logFile := flag.String("o", "", "путь к файлу логов")
	flag.StringVar(logFile, "out", "", "путь к файлу логов")
	errorFile := flag.String("e", "", "путь к файлу логов")
	flag.StringVar(errorFile, "errfile", "", "путь к файлу логов")
	adminka := flag.String("adminka", "", "Адрес админки. Если параметр не указан, то модуль не будет работать с админкой.")

	flag.Parse()

	cfgPath := *cfgFlag

	// установка временного префикса для системного журнала
	l.SetPrefix(moduleName)
	defer l.Close()
	l.SetVersion(2)

	l.Info("Запуск модуля:", moduleName)
	defer l.Info("Остановка модуля:", moduleName)

	if *appPort != -1 && *appSocket != "" {
		fmt.Printf("Одновременное использование -port и -socket не предусмотренно.\n")
		return
	}

	//Флаг для вывода ошибок Err Warning и panic обычной
	if *errorFile != "" {
		l.Info("errorFile=" + *errorFile)
		err := l.SetErrFileDesc(*errorFile)
		if err != nil {
			l.Err(11005, fmt.Sprintf("файл:%s, ошибка:%s\n", *errorFile, err.Error()))
			return
		}
		go l.ErrFileInit()
	}

	// работа с лок-файлом
	if !*testFlag {
		if lock, err = lockfile.Lock(*pidFlag); err != nil {
			l.Err(11001, err.Error())
			return
		}
		defer lock.Unlock()
		l.Debug("Установлена блокировка на PID файл")
	}
	// загрузка файловой конфигурации модуля
	if err = cfg.Load(cfgPath); err != nil {
		l.Err(10912, err.Error())
		return
	}

	// сериализациия конфигурации в JSON и вывод (получен -j флаг)
	if *jsonFlag {
		cfg.Print()
	}

	// проверка параметров полученных из командной строки
	var dsnArray []string
	dsnArray = append(dsnArray, getVar(cfg.Sql.PGUser, "PGUSER", *pgUser, "postgres://", ""))
	dsnArray = append(dsnArray, getVar(cfg.Sql.PGPassword, "PGPASSWORD", *pgPassword, ":", ""))
	dsnArray = append(dsnArray, getVar(cfg.Sql.PGHost, "PGHOST", *pgHost, "@", ""))
	dsnArray = append(dsnArray, getVar(cfg.Sql.PGPort, "PGPORT", *pgPort, ":", ""))
	dsnArray = append(dsnArray, getVar(cfg.Sql.PGDataBase, "PGDATABASE", *pgDataBase, "/", ""))
	dsnArray = append(dsnArray, getVar(cfg.Sql.PGFlags, "PGFLAGS", *pgFlags, "?", ""))
	tmpDSN := strings.Join(dsnArray, "")

	// перезапись значений представленных в конфигурации
	if tmpDSN == "" {
		if cfg.Sql.ConnectString != "" {
			l.Debug("Используем cfg.Sql.ConnectString из yaml-конфига")
		} else if cfg.Sql.ConnectString == "" && *dsn != "" {
			l.Debug("Используем -dsn из командной строки")
			cfg.Sql.ConnectString = *dsn
		}
	} else {
		cfg.Sql.ConnectString = tmpDSN
	}

	if *loglevel != -1 {
		cfg.LogLevel = *loglevel
	}
	if *journalName != "" {
		cfg.JournalName = *journalName
	}
	if *syslogFlag == "off" {
		cfg.Syslog = false
	}
	if *stdoutFlag == "off" {
		cfg.Stdout = false
	}
	if *dbPoolSizeFlag != -1 {
		cfg.Sql.PoolSize = *dbPoolSizeFlag
	}

	originLogLevel = cfg.LogLevel

	if *adminka != "" {
		cfg.Adminka = *adminka
	}

	if !cfg.Syslog {
		l.DisableSyslog()
	}

	if !cfg.Stdout {
		l.DisableLogStream("stderr")
	}

	// Устанавливаем количество иcпользуемых приложением процессоров
	if cfg.NumCpu > 0 {
		cpucores = cfg.NumCpu
	} else {
		cpucores = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(cpucores)

	// создаем пулл соединений с базой данных
	dbPool = *api.NewPool()
	if err = dbPool.FillPool(cfg.Sql.PoolSize, cfg.Sql.ConnectString); err != nil {
		l.Err(17001, err.Error())
		return
	}

	db, _ := dbPool.GetConn()
	// Задаем таймаут для запросов
	if *cfg.Sql.StatementTimeout > 0 {
		query := fmt.Sprintf("SET statement_timeout = %d;", *cfg.Sql.StatementTimeout)
		_, err = db.Exec(query)
		if err != nil {
			l.Err(17003, err.Error())
			return
		}
	}

	// запускаем крон, если там что-то есть
	c := cron.NewCron(db)

	// APPLIB.DataCtx - контект для связи *Inin функций
	// кладем туда соединение с базой
	APPLIB.DataCtx = context.WithValue(context.TODO(), "DB", db)
	APPLIB.DataCtx = context.WithValue(APPLIB.DataCtx, "Name", "dataCtx")

	if *errorFile != "" {
		//e := echo.New()
		APPLIB.DataCtx = context.WithValue(APPLIB.DataCtx, "ERRFILE", *errorFile)
	}

	APPLIB.DataCtx = context.WithValue(APPLIB.DataCtx, "CRON", c)
	APPLIB.DataCtx = context.WithValue(APPLIB.DataCtx, "DBPOOL", dbPool)
	ctx, CTXCancel := context.WithCancel(context.Background())
	APPLIB.DataCtx = context.WithValue(APPLIB.DataCtx, "CTX", ctx)

	// главный цикл (обеспечивает возможность перезапуска процесса)
	for reload {
		reload = false
		updCfgTime, err = config.LoadFromDb(&cfg, db)
		if err != nil {
			l.Err(11004, err.Error())
			return
		}
		// кладем в конектс конфиг (заполненный)
		APPLIB.DataCtx = context.WithValue(APPLIB.DataCtx, "CFG", cfg)

		hostname, err := os.Hostname()
		if err != nil {
			l.Err(12001, err.Error())
		}

		api.Meta = api.MetaType{
			PID:             os.Getpid(),
			CPU:             cpucores,
			Hostname:        hostname,
			Name:            moduleName,
			Release:         moduleVersion,
			GoTmplVersion:   goTemplateVersion,
			CompileTime:     moduleCompileTime,
			Commit:          gitCommit,
			InstanceGUID:    makeInstGUID(cfg, moduleName, hostname),
			CfgUpdTime:      updCfgTime.Format("2006-01-02T15:04:05"),
			CfgMD5Sum:       config.CfgMD5Sum,
			GoBuildVersion:  goBuildVersion,
			GoVersion:       goVersion,
			GoBuildTemplate: goBuildTemplate,
		}

		cfg.LogLevel = originLogLevel

		// настраиваем логер
		l.SetPrefix(cfg.JournalName)
		l.SetLogLevel(cfg.LogLevel)

		if *logFile != "" {
			l.Info("logFile=" + *logFile)
			err := l.SetLogFileDesc(*logFile)
			if err == nil {
				go l.LogFileInit()
			} else {
				l.Err(11005, fmt.Sprintf("файл:%s, ошибка:%s\n", *logFile, err.Error()))
			}
		}

		// валидация конфига и выход
		if *testFlag {
			return
		}

		// старт работы RMQ
		// кладем в дата-контекст объект пула
		if APPLIB.DataCtx, err = rmqInit(APPLIB.DataCtx); err != nil {
			l.Err("Ошибка в rmqInit:", err.Error())
			return
		}

		// старт работы крона
		if APPLIB.DataCtx, _, err = BG.CronInit(APPLIB.DataCtx); err != nil {
			l.Err("Ошибка при инициализации emp-go-cron:", err.Error())
			return
		}
		c.StartCron()

		// если при запуске передан порт, используем его вместо указанного в конфигурации
		if *appPort != -1 {
			cfg.Module.Http.Port = fmt.Sprintf(":%d", *appPort)
		}

		// если при запуске передан сокет, используем его вместо указанного в конфигурации
		if *appSocket != "" {
			cfg.Module.Http.UseSocket = true
			cfg.Module.Http.Socket = fmt.Sprintf("%s", *appSocket)
		}

		// Обработка сигналов
		signal.Notify(sigchan.SigChan,
			syscall.SIGINT,
			syscall.SIGILL,
			syscall.SIGHUP,
			syscall.SIGABRT,
			syscall.SIGBUS,
			syscall.SIGUSR1,
			syscall.SIGTERM,
			syscall.SIGTSTP,
			syscall.SIGCONT,
		)

		e, r := api.NewRouter(db, &cfg, api.Meta, sigchan.SigChan)
		APPLIB.DataCtx = context.WithValue(APPLIB.DataCtx, "ECHO", e)
		APPLIB.DataCtx = context.WithValue(APPLIB.DataCtx, "R", r)

		if APPLIB.DataCtx, err = APPLIB.ApplyRoutes(APPLIB.DataCtx); err != nil {
			l.Err("Ошибка в ApplyRoutes:", err.Error())
			return
		}

		go func() {
			l.Debug("Ожидание системных сигналов...")
			wg.Add(1)

			for {
				s := <-sigchan.SigChan
				switch s {
				case syscall.SIGINT, syscall.SIGTERM:
					l.Info("Выход. Получен сигнал SIGINT (0x2)")
					api.LogoutFromAdm(cfg.Adminka)
					ctx2, _ := context.WithTimeout(context.Background(), time.Duration(cfg.Module.Http.CancelingTimeOut)*time.Second)
					if err := e.Shutdown(ctx2); err != nil {
						if err := e.Close(); err != nil {
							l.Err(err)
						}
					}
					CTXCancel()
					dbPool.Close()
					l.OutFDesc.Close()
					APPLIB.TheEnd(APPLIB.DataCtx)
					time.Sleep(time.Second)
					c.DownCron()
					wg.Done()
					return

				case syscall.SIGUSR1:
					l.Info("Перезапуск. Получен сигнал SIGUSR1 (0xa)")
					api.LogoutFromAdm(cfg.Adminka)
					reload = true
					ctx2, _ := context.WithTimeout(context.Background(), time.Duration(cfg.Module.Http.CancelingTimeOut)*time.Second)
					if err := e.Shutdown(ctx2); err != nil {
						if err := e.Close(); err != nil {
							l.Err(err)
						}
					}
					CTXCancel()
					l.OutFDesc.Close()
					APPLIB.TheEnd(APPLIB.DataCtx)
					time.Sleep(time.Second)
					c.DownCron()
					wg.Done()
					return

				case syscall.SIGTSTP:
					l.Info("Установка уровня логирования на уровень ниже. Получен сигнал SIGTSTP (0x14)")
					if cfg.LogLevel == int(l.LOG_ERR) {
						l.Info("Уже установлен минимальный уровень логирования")
						continue
					}
					cfg.LogLevel = cfg.LogLevel - 1
					l.SetLogLevel(cfg.LogLevel)
					l.Info(fmt.Sprintf("Установлен уровень логирования: %s (%d)", l.GetTextLogLevel(), l.GetLogLevel()))

				case syscall.SIGCONT:
					l.Info("Установка уровня логирования на уровень выше. Получен сигнал SIGCONT (0x12)")
					if cfg.LogLevel == int(l.LOG_DEBUG) {
						l.Info("Уже установлен максимальный уровень логирования")
						continue
					}
					cfg.LogLevel = cfg.LogLevel + 1
					l.SetLogLevel(cfg.LogLevel)
					l.Info(fmt.Sprintf("Установлен уровень логирования: %s (%d)", l.GetTextLogLevel(), l.GetLogLevel()))

				case syscall.SIGILL:
					l.Info("Установка уровня логирования: LOG_ERR (3). Получен сигнал SIGILL (0x4)")
					cfg.LogLevel = int(l.LOG_ERR)
					l.SetLogLevel(cfg.LogLevel)

				case syscall.SIGABRT:
					l.Info("Установка уровня логирования: LOG_INFO (6). Получен сигнал SIGABRT (0x6)")
					cfg.LogLevel = int(l.LOG_INFO)
					l.SetLogLevel(cfg.LogLevel)

				case syscall.SIGBUS:
					l.Info("Установка уровня логирования: LOG_DEBUG (7). Получен сигнал SIGBUS (0x7)")
					cfg.LogLevel = int(l.LOG_DEBUG)
					l.SetLogLevel(cfg.LogLevel)

				case syscall.SIGHUP:
					l.Info("Получен сигнал SIGHUP (0x1). Пробуем переоткрыть файл логов")
					if *logFile != "" {
						//открываем и закрываем файл
						if l.ErrFileDesc = l.OutFDesc.Close(); l.ErrFileDesc != nil {
							l.Err("Ошибка закрытия лог файла (-o): " + l.ErrFileDesc.Error())
						}
						// TODO: обработать ошибку
						l.SetLogFileDesc(*logFile)
					}

					l.Info("Получен сигнал SIGHUP (0x1). Пробуем переоткрыть файл ошибок panic, l.Err, l.Warn")
					if *errorFile != "" {
						//открываем и закрываем файл
						if l.FileError = l.ErrFDesc.Close(); l.FileError != nil {
							l.Err("Ошибка закрытия лог файла (-e): " + l.FileError.Error())
						}
						// TODO: обработать ошибку
						l.SetErrFileDesc(*errorFile)
					}
				default:
					// В эту часть кода мы никогда не попадем...
					l.Info("Получен необрабатываемый тип сигнала: %d", s)
				}
			}
		}()

		// добавляем middleware проверки соединения с БД
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// проверяем состояние пулла и если там проблемы, то переконнект
				var (
					err error
				)
				dbPool, err = checkConnect(dbPool, cfg)
				if err != nil {
					return err
				}

				if err := next(c); err != nil {
					return err
				}
				return nil
			}
		})

		// вызов функции, настраивающей миддлварь для эхо
		if APPLIB.DataCtx, err = echoInit(APPLIB.DataCtx); err != nil {
			l.Err(10911, err.Error())
			return
		}

		fmt.Printf("%s", instance.STAT)

		if cfg.Module.Http.UseSocket {
			os.Remove(cfg.Module.Http.Socket)
			listener, err := net.Listen("unix", cfg.Module.Http.Socket)
			if err != nil {
				l.Err("WebServer не смог открыть сокет")
				return
			}
			if err = os.Chmod(cfg.Module.Http.Socket, 0760); err != nil {
				l.Err("Ошибка при смене прав на сокет")
				return
			}

			e.Listener = listener
			l.Info("WebServer запуcкается ", cfg.Module.Http.Socket)
		} else {
			netport = cfg.Module.Http.Port
			l.Info("WebServer запуcкается ", netport)
		}

		if err := e.Start(netport); err != nil && err.Error() == errServClosed {
			l.Info("WebServer выключается...")
		} else if err != nil {
			l.Err(14911, "Ошибка при запуске веб-сервера", err.Error())
			return
		}

		wg.Wait()
	}
}

// getVar - получение данных для соедиения с Базой данных
// Важно, если ни одно из трех первых переменных даст результата, то префиск и суффикс использованы не будут
// т.е. getVar("","","","A","B") => "", пустая строка, а не "AB"
// при этом getVar("и","","","A","B") => "AиB"
//		s -данные из конфига
//		envKey - имя переменной окружения
//		cli - парамет командной строки
//		prevStr - пекифкс
// 		suffixStr - суффикс
func getVar(cfgData, envKey, cli, prevStr, suffixStr string) string {
	if cli != "" {
		l.Debug("Переменная " + envKey + " найдена в командной строке")
		return prevStr + cli + suffixStr
	}

	if cfgData != "" {
		l.Debug("Переменная " + envKey + " найдена в конфиге")
		return prevStr + cfgData + suffixStr
	}

	ret := os.Getenv(envKey)
	if ret != "" {
		l.Debug("Переменная " + envKey + " найдена в окружении")
		return prevStr + ret + suffixStr
	}
	l.Err(10910, "Переменная "+envKey+" не найдена")
	return ""
}

func makeInstGUID(cfg config.Config, modulename, hostname string) string {
	ps := cfg.Module.Http.Port
	if cfg.Module.Http.UseSocket {
		ps = cfg.Module.Http.Socket
	}

	return common.MakeMD5String("", modulename, hostname, ps)

}

// checkConnect - проверка соедиения с БД
func checkConnect(p api.DBPool, cfg config.Config) (api.DBPool, error) {
	var (
		err error
	)
	db, _ := p.GetConn()
	if err = db.Ping(); err != nil {
		if err = p.FillPool(cfg.Sql.PoolSize, cfg.Sql.ConnectString); err != nil {
			return p, err
		}
	}

	return p, nil
}
