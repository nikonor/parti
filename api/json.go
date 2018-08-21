package api

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/labstack/echo"
	"gopkg.in/go-playground/validator.v9"

	l "parti/logger"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

const (
	//Err403str текст ошибки 403
	Err403str = "Пользователь с такой парой login/password не найден"
	//Err405str текст ошибки 405
	Err405str = "Параметр %s является обязательным"
	// Err406str тест ошибки 406
	Err406str = "Параметр %s имеет неверный формат"
	// Err408str тест ошибки 408
	Err408str = "Параметр %s имеет неверное значение"
	// TimeFormat Формат даты
	TimeFormat = "2006-01-02 15:04:05-07:00"
	// OutTimeFormat Формат даты для выдачи наружу
	OutTimeFormat = " 2006-01-02 15:04:05"
)
const (
	InstanceStatusOff = iota
	InstanceStatusOn
	InstanceStatusWillReload
	InstanceStatusServerMode
)

// EmpNullString null строка
type EmpNullString sql.NullString

// UnmarshalJSON декодирует значение в структуру, хранящую null значение
func (e *EmpNullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		e.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &e.String); err != nil {
		fmt.Println(err.Error())
		return err
	}
	e.Valid = true
	return nil
}

// Value реализует интерфейс Valuer драйвера
func (e EmpNullString) Value() (driver.Value, error) {
	if !e.Valid {
		return nil, nil
	}
	return e.String, nil
}

// EmpNullInt64 null int64
type EmpNullInt64 sql.NullInt64

// UnmarshalJSON декодирует значение в структуру, хранящую null значение
func (e *EmpNullInt64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		e.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &e.Int64); err != nil {
		return err
	}
	e.Valid = true
	return nil
}

// Value реализует интерфейс Valuer драйвера
func (e EmpNullInt64) Value() (driver.Value, error) {
	if !e.Valid {
		return nil, nil
	}
	return e.Int64, nil
}

// EmpNullBool null bool
type EmpNullBool sql.NullBool

// UnmarshalJSON декодирует значение в структуру, хранящую null значение
func (e *EmpNullBool) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		e.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &e.Bool); err != nil {
		return err
	}
	e.Valid = true
	return nil
}

// Value реализует интерфейс Valuer драйвера
func (e EmpNullBool) Value() (driver.Value, error) {
	if !e.Valid {
		return nil, nil
	}
	return e.Bool, nil
}

// EmpNullFloat64 null float64
type EmpNullFloat64 sql.NullFloat64

// UnmarshalJSON декодирует значение в структуру, хранящую null значение
func (e *EmpNullFloat64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		e.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &e.Float64); err != nil {
		return err
	}
	e.Valid = true
	return nil
}

// Value реализует интерфейс Valuer драйвера
func (e EmpNullFloat64) Value() (driver.Value, error) {
	if !e.Valid {
		return nil, nil
	}
	return e.Float64, nil
}

// AccessToDB - интерефейс, дающий возможность работать как через sql.DB, так и через sql.Tx
type AccessToDB interface {
	QueryRow(query string, argv ...interface{}) *sql.Row
	Query(query string, argv ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
}

// Container - струтура для стандартного ответа
type Container struct {
	ErrorCode    int         `json:"errorCode" comment:"0, если все прошло хорошо. И не 0, если что-то пошло не так."`
	ErrorMessage string      `json:"errorMessage" comment:"Сообщение об ошибке. Заполняется только если errorCode не равно 0"`
	Result       interface{} `json:"result" comment:"Массив, в который кладется значимая информация ответа"`
}

//LoginResponse - структура ответа о регистрации инстанса
type LoginResponse struct {
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	Result       []struct {
		InstanceID  int `json:"instance_id"`
		OrderNumber int `json:"order_number"`
	} `json:"result"`
}

// ErrorHandler - обработка ошибок
func ErrorHandler(e error, c echo.Context) {
	code := 500
	mes := "Внутренняя ошибка"
	ee, ok := e.(*echo.HTTPError)
	if ok {
		code = ee.Code
		mes = ee.Message.(string)
	}

	stack := make([]byte, 16<<10)
	length := runtime.Stack(stack, true)

	panicEcho := string(stack[:length])
	//отлов паники echo - фреймворка
	if strings.Contains(panicEcho, "panic") {
		l.Err(code, e.Error()+": "+string(stack[:length]))
	} else {
		l.Err(code, c.Path()+" "+mes)
	}

	Fault(c, code, mes)
}

// Response - обертка для возврата ответа от API в стандартном контейнере
func Response(c echo.Context, result interface{}, headers ...map[string]string) error {
	l.Debug("Вызов api.Response")

	var (
		err     error
		payload Container
	)

	if len(headers) > 0 {
		for k, v := range headers[0] {
			c.Response().Header().Add(k, v)
		}
	}

	payload.ErrorCode = 0
	payload.ErrorMessage = ""

	// fmt.Println(reflect.TypeOf(result).Kind())

	switch reflect.TypeOf(result).Kind() {
	case reflect.Slice, reflect.Array:
		payload.Result = result

	default: // t, _ := reflect.TypeOf(in).FieldByName(err.(validator.ValidationErrors)[0].Field())
		// return false, strings.Replace(t.Tag.Get("json"), ",omitempty", "", -1)

		payload.Result = []interface{}{result}
	}

	if err = c.JSON(http.StatusOK, payload); err != nil {
		message, code := l.Err(14903, err.Error())
		return Fault(c, code, message)
	}

	return nil
}

// ProxyResponse - обертка для проксированного ответа
func ProxyResponse(c echo.Context, result []byte, onlyProxy bool, headers ...map[string]string) error {
	l.Debug("Вызов api.ProxyResponse")

	var (
		err     error
		payload []byte
		flag    bool
	)

	if len(headers) > 0 {
		for k, v := range headers[0] {
			c.Response().Header().Add(k, v)
		}
	}
	if result[0] != '[' {
		flag = true
	}
	if !onlyProxy {
		payload = append(payload, []byte(`{"errorCode":0,"errorMessage":"","result":`)...)
		if flag {
			payload = append(payload, '[')
		}
	}
	payload = append(payload, result...)
	if !onlyProxy {
		if flag {
			payload = append(payload, ']')
		}
		payload = append(payload, []byte(`}`)...)
	}

	if err = c.JSONBlob(http.StatusOK, payload); err != nil {
		message, code := l.Err(14904, err.Error())
		return Fault(c, code, message)
	}

	return nil
}

// BlobResponse - обертка для проксированного ответа
func BlobResponse(c echo.Context, contentType string, body []byte, headers ...map[string]string) error {
	l.Debug("Вызов api.BlobResponse")

	var (
		err error
	)

	if len(headers) > 0 {
		for k, v := range headers[0] {
			c.Response().Header().Add(k, v)
		}
	}

	if err = c.Blob(http.StatusOK, contentType, body); err != nil {
		message, code := l.Err(14905, err.Error())
		return Fault(c, code, message)
	}

	return nil
}

// Fault - обертка для возврата ошибки
func Fault(c echo.Context, code int, message string, headers ...map[string]string) error {

	s, ec := l.Err(code, message)
	if s == "" {
		s = message
	}
	if ec == 0 {
		ec = code
	}

	var payload Container
	if len(headers) > 0 {
		for k, v := range headers[0] {
			c.Response().Header().Add(k, v)
		}
	}

	payload.ErrorCode = ec
	payload.ErrorMessage = s
	payload.Result = []interface{}{}

	return c.JSON(http.StatusOK, payload)
}

// FaultF - обертка для возврата ошибки
func FaultF(c echo.Context, code int, message string, additionToPublicMessage []string, headers ...map[string]string) error {

	s, ec := l.Err(code, message)
	if s == "" {
		s = message
	}
	if ec == 0 {
		ec = code
	}

	for i, aStr := range additionToPublicMessage {
		s = strings.Replace(s, fmt.Sprintf("$%d", i), aStr, 1)
	}

	var payload Container
	if len(headers) > 0 {
		for k, v := range headers[0] {
			c.Response().Header().Add(k, v)
		}
	}

	payload.ErrorCode = ec
	payload.ErrorMessage = s
	payload.Result = []interface{}{}

	return c.JSON(http.StatusOK, payload)
}

// Ok - обертка для возврата пустого успешного ответа
func Ok(c echo.Context, headers ...map[string]string) error {
	l.Debug("Вызов api.Ok")

	if len(headers) > 0 {
		for k, v := range headers[0] {
			c.Response().Header().Add(k, v)
		}
	} // 	pp := strings.Split(path, ".")

	payload := Container{
		ErrorCode:    0,
		ErrorMessage: "",
	}

	return c.JSON(http.StatusOK, payload)
}

// DecodeWrapper - обертка над json.Decoder.Decode
func DecodeWrapper(dec *json.Decoder, out interface{}) (int, string) {
	if err := dec.Decode(&out); err != nil {
		switch err.(type) {
		case *json.UnmarshalTypeError:
			return 406, "Параметр " + err.(*json.UnmarshalTypeError).Field + " имеет неверный формат"
		default:
			return 400, "Тело запроса содержит некорректную структуру JSON"
		}
	}

	return 0, ""
}

// DecodeBytes - декодирование []bytes с проверками
func DecodeBytes(in []byte, out interface{}) (int, string) {
	dec := json.NewDecoder(bytes.NewReader(in))
	return DecodeWrapper(dec, out)
}

// DecodeReader - декодирование io.Reader с проверками
func DecodeReader(in io.Reader, out interface{}) (int, string) {
	dec := json.NewDecoder(in)
	return DecodeWrapper(dec, out)
}

// Decode - функция декодирования полученного запроса
//	in м.б. либо []byte, либо io.Reader
func Decode(in interface{}, out interface{}) (int, string) {
	switch in.(type) {
	case []byte:
		return DecodeBytes(in.([]byte), out)
	case io.Reader:
		return DecodeReader(in.(io.Reader), out)
	}
	return -1, ""
}

// Validator - провера заполненности полей
func Validator(in interface{}) (bool, int, string) {
	var (
		err     error
		errCode = 405
	)
	v := validator.New()

	// функия для проверик на required для sql.Null*, объявленных через EmpNull* типы
	v.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if field.FieldByName("Valid").Bool() {
			if reflect.TypeOf(field.Interface()).Name() == "EmpNullString" {
				return field.FieldByName("String")
			}
			if reflect.TypeOf(field.Interface()).Name() == "EmpNullInt64" {
				return field.FieldByName("Int64")
			}
			if reflect.TypeOf(field.Interface()).Name() == "EmpNullBool" {
				return field.FieldByName("Bool")
			}
			if reflect.TypeOf(field.Interface()).Name() == "EmpNullFloat64" {
				return field.FieldByName("Float64")
			}
		}
		return nil
	}, EmpNullBool{}, EmpNullFloat64{}, EmpNullInt64{}, EmpNullString{})

	// функция для получения JSON нотации в случае ошибки.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}
		return name
	})

	if err = v.Struct(in); err != nil {
		// fmt.Println("Namespace=", err.(validator.ValidationErrors)[0].Namespace())
		// fmt.Println("Field=", err.(validator.ValidationErrors)[0].Field())
		// fmt.Println("StructNamespace())=", err.(validator.ValidationErrors)[0].StructNamespace())
		// fmt.Println("StructField())=", err.(validator.ValidationErrors)[0].StructField())
		// fmt.Println("Tag=", err.(validator.ValidationErrors)[0].Tag())
		// fmt.Println("ActualTag=", err.(validator.ValidationErrors)[0].ActualTag())
		// fmt.Println("Kind=", err.(validator.ValidationErrors)[0].Kind())
		// fmt.Println("Type=", err.(validator.ValidationErrors)[0].Type())
		// fmt.Println("Value=", err.(validator.ValidationErrors)[0].Value())
		// fmt.Println("Param=", err.(validator.ValidationErrors)[0].Param())
		switch err.(validator.ValidationErrors)[0].ActualTag() {
		case "required":
			errCode = 405
		case "email":
			errCode = 406
		case "eq", "gt", "gte", "lte":
			errCode = 406
		default:
			errCode = 406
		}

		return false, errCode, strings.Join(strings.Split(err.(validator.ValidationErrors)[0].Namespace(), ".")[1:], ".")
		// return false, err.(validator.ValidationErrors)[0].StructNamespace()
		// return false, errCode, getJSONTagsError(in, err.(validator.ValidationErrors)[0].StructNamespace())
	}

	return true, 0, ""
}

func getJSONTagsError(i interface{}, path string) string {
	println(path)
	var (
		ret string
		in  reflect.Value
	)

	switch i.(type) {
	case reflect.Value:
		in = i.(reflect.Value)
	default:
		in = reflect.Indirect(reflect.ValueOf(i))
	}
	pp := strings.Split(path, ".")
	if len(pp) > 1 {
		ff := strings.Split(pp[1], "[")
		fieldName := ff[0]
		if in.Kind() == reflect.Ptr {
			in = reflect.Indirect(in)
		}
		typ := reflect.TypeOf(in.Interface())
		if t, ok := typ.FieldByName(fieldName); ok {
			ret = strings.Replace(t.Tag.Get("json"), ",omitempty", "", 1)
		}

		idx := int64(0)
		if len(ff) == 2 {
			idx, _ = strconv.ParseInt(strings.Replace(ff[1], "]", "", 1), 10, 64)
			ret = fmt.Sprintf("%s[%d]", ret, idx)
		}

		nextIn := reflect.ValueOf(in.Interface()).FieldByName(fieldName)
		switch nextIn.Kind() {
		case reflect.Slice, reflect.Array:
			if nextIn.Len() == 0 {
				return ret
			}
			nextIn = nextIn.Index(int(idx))

		}
		add := getJSONTagsError(nextIn, strings.Join(pp[1:], "."))
		if len(add) > 0 {
			ret = fmt.Sprintf("%s.%s", ret, add)
		}
	}

	return ret
}

// UniversalBegin - универсальное начало для всех функций
func UniversalBegin(in, out interface{}) (int, string) {
	var (
		errCode    int
		errMessage string
		ok         bool
		body       []byte
	)

	switch in.(type) {
	case echo.Context:
		body, _ = ioutil.ReadAll(in.(echo.Context).Request().Body)
	case http.Response:
		body, _ = ioutil.ReadAll(in.(http.Response).Body)
	case *http.Response:
		body, _ = ioutil.ReadAll(in.(*http.Response).Body)
	case []byte:
		body = in.([]byte)
	}

	l.Debug(string(body))

	if errCode, errMessage = Decode(body, &out); errCode > 0 {
		return errCode, errMessage
	}
	// валидация
	if ok, errCode, errMessage = Validator(out); !ok {
		tmpl := Err405str
		if errCode != 405 {
			tmpl = Err406str
		}
		return errCode, strings.Replace(tmpl, "%s", errMessage, 1)
	}

	return 0, ""
}

// UniversalBeginMasked - универсальное начало для всех функций
func UniversalBeginMasked(in, out interface{}) (int, string) {
	var (
		errCode    int
		errMessage string
		ok         bool
		body       []byte
	)

	switch in.(type) {
	case echo.Context:
		body, _ = ioutil.ReadAll(in.(echo.Context).Request().Body)
	case http.Response:
		body, _ = ioutil.ReadAll(in.(http.Response).Body)
	case *http.Response:
		body, _ = ioutil.ReadAll(in.(*http.Response).Body)
	case []byte:
		body = in.([]byte)
	}

	// TODO: сделать замену пароля на ***?
	// l.Debug(string(body))

	if errCode, errMessage = Decode(body, &out); errCode > 0 {
		return errCode, errMessage
	}
	// валидация
	if ok, errCode, errMessage = Validator(out); !ok {
		tmpl := Err405str
		if errCode != 405 {
			tmpl = Err406str
		}
		return errCode, strings.Replace(tmpl, "%s", errMessage, 1)
	}

	return 0, ""
}
