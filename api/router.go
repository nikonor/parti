package api

import "github.com/labstack/echo"
import "github.com/labstack/echo/middleware"

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"parti/config"
	l "parti/logger"
	"parti/sigchan"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	// Tab - отступ для каждого уровня в документации
	Tab = "  "
	// Meta - набор метаданных о модуле
	Meta      MetaType
	StartTime time.Time
)

type MetaType struct {
	PID             int    `json:"pid"`
	CPU             int    `json:"cpu"`
	Hostname        string `json:"host"`
	Name            string `json:"name"`
	Release         string `json:"version"`
	Commit          string `json:"commit"`
	CompileTime     string `json:"build_time"`
	GoTmplVersion   string `json:"go_template_version"`
	InstanceGUID    string `json:"instance_guid"`
	CfgMD5Sum       string `json:"cfg_md5_sum"`
	CfgUpdTime      string `json:"cfg_update_time"`
	GoBuildVersion  string `json:"go_build_version"`
	GoVersion       string `json:"go_version"`
	GoBuildTemplate string `json:"go_build_template"`
	UpTime          string `json:"up_time"`
}

// IOField ...
type IOField struct {
	Name     string     `json:"name"`
	Type     string     `json:"type"`
	Required bool       `json:"required"`
	Desc     string     `json:"description"`
	Link     []*IOField `json:"properties,omitempty"`
}

// IOStruct стуктура для хранения описания структур
type IOStruct struct {
	Name   string     `json:"name,omitempty"`
	Fields []*IOField `json:"fields,omitempty"`
}

// CommentType структура для хранения комментария
type CommentType struct {
	Method        string      `json:"method,omitempty"`
	Path          string      `json:"path,omitempty"`
	Desc          string      `json:"title"`
	Description   string      `json:"description"`
	InputStructs  []*IOStruct `json:"input_structs,omitempty"`
	OutputStructs []*IOStruct `json:"output_structs,omitempty"`
}

// Route - элемент роутинга
type Route struct {
	Handler   echo.HandlerFunc
	Regexp    *regexp.Regexp
	Method    string
	Path      string
	Comment   string
	JSONHelp  []byte
	Options   map[string]bool
	ReqFields []string
}

type Key string

// SortedKeysType тип для сортировки вывода
type SortedKeysType []string

func (a SortedKeysType) Len() int           { return len(a) }
func (a SortedKeysType) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortedKeysType) Less(i, j int) bool { return a[i] < a[j] }

func swapKey(s string) string {
	return strings.Join(strings.Split(s, " "), " ")
}

// RouterHandlers - ?
type RouterHandlers struct {
	handlers map[string]Route
}

// NewRouter - создание новой схему роутинга запросов
func NewRouter(db *sql.DB, cfg *config.Config, meta MetaType, sigChan chan os.Signal) (*echo.Echo, *RouterHandlers) {
	r := RouterHandlers{handlers: make(map[string]Route)}

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = ErrorHandler
	//e.Use(middleware.Logger())
	rc := middleware.RecoverConfig{
		StackSize: 16 << 10, // 16Kb @TODO:move to config
		//DisablePrintStack: true,
	}
	e.Use(middleware.RecoverWithConfig(rc))

	// корневой роут - ф-я мониторинга
	// e.GET("/", func(c echo.Context) error {
	// 	meta.UpTime = time.Since(StartTime).String()
	// 	c.JSON(http.StatusOK, meta)
	// 	return nil
	// })
	e.GET("/m", func(c echo.Context) error {
		meta.UpTime = time.Since(StartTime).String()
		fmt.Println("meta = %#v", meta)
		c.JSON(http.StatusOK, meta)
		return nil
	})
	e.GET("/m/", func(c echo.Context) error {
		meta.UpTime = time.Since(StartTime).String()
		c.JSON(http.StatusOK, meta)
		return nil
	})

	// вернуть конфигурацию модуля
	e.GET("/m/config/get", func(c echo.Context) error {
		conf, err := config.GetHandler(cfg, db)
		c.Response().Header().Add("content-type", "application/json")
		c.Response().Write(conf)
		return err
	})

	// редактировать конфигурацию модуля
	e.POST("/m/config/set", func(c echo.Context) error {
		err := config.SaveHandler(c.Request().Body, cfg, db)
		if err != nil {
			return Fault(c, 500, err.Error())
		}
		// @TODO: обработка ошибки..
		return OkHandler(c)
	})

	// вернуть схему конфигурации
	e.GET("/m/config/schema", func(c echo.Context) error {
		schema, err := config.SchemaHandler()
		c.Response().Write(schema)
		return err
	})

	e.POST("/m/app/shutdown", func(c echo.Context) error {
		sigchan.ShutDown()
		return nil
	})

	// перезагрузить модуль @TODO: добавить задержку в секундах.
	e.POST("/m/app/reload", func(c echo.Context) error {
		sigchan.Reload()
		return nil
	})

	// автодокументация - plaintext форматирование
	e.GET("/m/doc/txt", func(c echo.Context) error {
		txt := r.docTxt()
		c.Response().Write(txt)
		return nil
	})

	e.POST("/m/file/close", func(c echo.Context) error {
		l.WriteLogFile <- false
		return nil
	})

	e.POST("/m/file/open", func(c echo.Context) error {
		l.WriteLogFile <- true
		return nil
	})

	// автодокументация - json форматирование
	e.GET("/m/doc/json", func(c echo.Context) error {
		json := r.docJson()
		c.Response().Header().Add("content-type", "application/json")
		c.Response().Write([]byte(`{"errorCode":0, "errorMessage": "", "result":[`))
		c.Response().Write(bytes.Join(json, []byte(",")))
		c.Response().Write([]byte("]}"))
		return nil
	})

	return e, &r
}

func GET(
	e *echo.Echo,
	r *RouterHandlers,
	path string,
	h echo.HandlerFunc,
	in []interface{},
	out []interface{},
	title string,
	description string,
	m ...echo.MiddlewareFunc,
) *echo.Route {
	return Add(e, r, "GET", path, h, in, out, title, description, m...)
}

func POST(
	e *echo.Echo,
	r *RouterHandlers,
	path string,
	h echo.HandlerFunc,
	in []interface{},
	out []interface{},
	title string,
	description string,
	m ...echo.MiddlewareFunc,
) *echo.Route {
	return Add(e, r, "POST", path, h, in, out, title, description, m...)
}

func Add(
	e *echo.Echo,
	r *RouterHandlers,
	method string,
	path string,
	h echo.HandlerFunc,
	in []interface{},
	out []interface{},
	title string,
	description string,
	m ...echo.MiddlewareFunc,
) *echo.Route {

	var (
		re  *regexp.Regexp
		key string
		j   []byte
	)

	echoRoute := e.Add(method, path, h, m...)

	_, path, key = makeMask(method, path)
	key = path + " " + method
	if _, ok := r.handlers[key]; ok {
		//@TODO: log error: Ошибка добавления маршрута %s %s: не уникальный ", method, path
	}

	// работаем с help-ом
	jsonComment := CommentType{
		Desc:        title,
		Method:      method,
		Path:        path,
		Description: description,
	}

	if title != "-" {
		if len(in) > 0 {
			title = title + "\n" + Tab + "Структуры запроса" + ":\n"
			for _, o := range in {
				s, ss := getDopInfoForJSON(o, 2)
				title = title + ss + "\n"
				jsonComment.InputStructs = append(jsonComment.InputStructs, &s)
			}
			jsonComment.InputStructs = glueStructs(jsonComment.InputStructs)
		}

		if len(out) > 0 {
			title = title + "\n" + Tab + "Структуры ответа" + ":\n"
			for _, o := range out {
				s, ss := getDopInfoForJSON(o, 2)
				title = title + ss + "\n"
				jsonComment.OutputStructs = append(jsonComment.OutputStructs, &s)
			}
			jsonComment.OutputStructs = glueStructs(jsonComment.OutputStructs)
		}
		j, _ = json.Marshal(jsonComment)
	}

	if description != "" {
		title = title + "\n" + Tab + description
	}

	r.handlers[key] = Route{
		Handler:  h,
		Method:   method,
		Path:     path,
		Regexp:   re,
		Comment:  title,
		JSONHelp: j,
	}

	//fmt.Printf("%#v", r)

	return echoRoute
}

// OkHandler - тестовый обработчик, успешный ответ
func OkHandler(c echo.Context) error {
	type R struct {
		Now time.Time `json:"now"`
	}
	data := R{Now: time.Now()}
	return Response(c, data)
}

func (r RouterHandlers) docJson() [][]byte {
	var (
		sortedKeys SortedKeysType
		buffer     [][]byte
	)

	type EndPoint struct {
		Method, Path, Comment string
	}

	for k := range r.handlers {
		sortedKeys = append(sortedKeys, k)
	}

	sort.Sort(sortedKeys)
	for _, k := range sortedKeys {
		if r.handlers[k].Comment != "-" {
			buffer = append(buffer, r.handlers[k].JSONHelp)
		}
	}
	//fmt.Printf("%#v", buffer)
	return buffer
}

func (r RouterHandlers) docTxt() []byte {

	var (
		sortedKeys SortedKeysType
	)

	type EndPoint struct {
		Method, Path, Comment string
	}

	ret := ""

	for k := range r.handlers {
		sortedKeys = append(sortedKeys, k)
	}

	sort.Sort(sortedKeys)

	for _, k := range sortedKeys {
		k = swapKey(k)
		h := r.handlers[k]
		if h.Comment != "-" {
			c := ""
			for i, s := range strings.Split(h.Comment, "\n") {
				if i == 0 {
					c = c + fmt.Sprintf("%s\n", s)
				} else {
					c = c + fmt.Sprintf("%s%s\n", Tab, s)
				}
			}

			ret = ret + fmt.Sprintf("Вызов: %s %s\n%s%s\n", h.Method, h.Path, Tab, c)
		}
	}

	return []byte(ret)
}

func makeMask(method, path string) (string, string, string) {
	method = strings.ToUpper(method)
	path = strings.TrimRight(path, "/")
	if method == "*" {
		method = `\w+`
	}
	key := method + " " + path

	path = strings.Replace(path, `(\d+)`, `[0-9]`, -1)
	path = strings.Replace(path, `(\w+)`, `[a-zA-Z]`, -1)

	return method, path, key
}

func glueStructs(objects []*IOStruct) []*IOStruct {
	type Element struct {
		Obj     *IOStruct
		IsChild bool
	}
	var (
		ret []*IOStruct
		mm  map[string]*Element
	)
	mm = make(map[string]*Element)
	// Шаг 1 - создаем словарь объектов
	for _, oOuter := range objects {
		mm[oOuter.Name] = &Element{Obj: oOuter}
	}
	// Шаг 2 - бежим по объектам
	for _, oOuter := range objects {
		// В каждом объекте бежим по полям
		for _, f := range oOuter.Fields {
			if strings.Contains(f.Type, ".") {
				for k, o := range mm {
					if k != oOuter.Name && strings.HasSuffix(f.Type, k) {
						// if k != oOuter.Name && strings.Contains(f.Type, k) {
						// если нашли совпадение, то связываем поля
						o.IsChild = true
						f.Link = o.Obj.Fields
					}
				}
			}
		}
	}
	// Шаг 3 - прореживаем набор
	for _, o := range objects {
		if !mm[o.Name].IsChild {
			ret = append(ret, o)
		}
	}

	return ret
}

func getDopInfoForJSON(o interface{}, level int) (IOStruct, string) {
	var (
		ret    IOStruct
		retStr string
		tab    string
	)

	for i := 0; i < level; i++ {
		tab = fmt.Sprintf("%s%s", tab, Tab)
	}

	val := reflect.Indirect(reflect.ValueOf(o))
	t := val.Type()
	retStr = tab + "Структура " + t.String() + "\n"
	tab = tab + Tab
	ret.Name = t.String()
	for i := 0; i < t.NumField(); i++ {
		f, fStr := getFieldCommentForJSON(t.Field(i))
		if f != nil {
			ret.Fields = append(ret.Fields, f)
		}
		if fStr != "" {
			retStr = retStr + fmt.Sprintf("%s%s\n", tab, fStr)
		}

	}
	return ret, retStr
}

func getFieldCommentForJSON(f reflect.StructField) (*IOField, string) {

	r := IOField{}
	rStr := ""

	tComment := f.Tag.Get("comment")
	if tComment == "-" {
		return nil, ""
	}
	tJSON := strings.Replace(f.Tag.Get("json"), ",omitempty", "", 1)
	if tJSON == "-" {
		return nil, ""
	}
	if tJSON != "" {
		r.Name = tJSON
		rStr = tJSON
	} else {
		r.Name = f.Name
		rStr = f.Name
	}
	r.Type = fmt.Sprintf("%s", f.Type)
	rStr = rStr + "(" + fmt.Sprintf("%s", f.Type) + ")"

	if tComment != "" {
		r.Desc = tComment
		rStr = rStr + ": " + tComment
	}

	// работаем с required
	tReq := f.Tag.Get("validate")
	if len(tReq) >= 1 && strings.Contains(tReq, "required") {
		r.Required = true
		rStr = rStr + ". ОБЯЗАТЕЛЬНОЕ"
	}

	return &r, rStr
}
