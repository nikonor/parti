package config

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"

	l "parti/logger"
	"io"
	"io/ioutil"
	"strings"
)

var (
	// CfgMD5Sum - md5 сумма конфига
	CfgMD5Sum string
)

// configReq структура http запроса содержащего конфиг
type configReq struct {
	Name        string `json:"name"`
	Data        Config `json:"data"`
	Description string `json:"description"`
}

// configResp структура http ответа содержащего конфиг
type configResp struct {
	ErrorCode    int     `json:"errorCode"`
	ErrorMessage string  `json:"errorMessage"`
	Result       *Config `json:"result,omitempty"`
}

// schemaResp структура http ответа содержащего JSON схему
type schemaResp struct {
	ErrorCode    int             `json:"errorCode"`
	ErrorMessage string          `json:"errorMessage"`
	Result       json.RawMessage `json:"result,omitempty"`
}

// loaders - подготовка перед валидацией конфигурации из СУБД по схеме
func loaders(data sql.NullString) (gojsonschema.JSONLoader, gojsonschema.JSONLoader) {
	l.Debug("Вызов config.loaders")
	schemaLoader := gojsonschema.NewStringLoader(SQLConfigSchema)
	documentLoader := gojsonschema.NewStringLoader(data.String)
	return schemaLoader, documentLoader
}

// Load - загрузка файловой конфигурации модуля
func (cfg *Config) Load(path string) error {
	yamlData, err := ioutil.ReadFile(path)
	if err != nil {
		l.Err(11002, err.Error())
		return err
	}

	// десериализация YAML конфигурации
	err = yaml.Unmarshal(yamlData, &cfg)
	if err != nil {
		l.Err(11003, err.Error())
		return err
	}

	// проверка файлового конфига модуля
	schemaLoader := gojsonschema.NewStringLoader(YAMLConfigSchema)
	documentLoader := gojsonschema.NewGoLoader(&cfg)

	return Check("Файловая конфигурация модуля", schemaLoader, documentLoader)
}

// Print - вывод конфигурации в формате json
func (cfg *Config) Print() {
	configJSON, _ := json.Marshal(cfg)
	fmt.Printf("Конфигурация в формате json:\n%s\n", string(configJSON))
}

// LoadFromDb - загрузка конфига из базы данных
func LoadFromDb(conf *Config, db *sql.DB) (time.Time, error) {

	var (
		updCfgTime time.Time
		err        error
		data       sql.NullString
	)

	l.Debug("Загрузка конфигурации модуля...")
	defer l.Debug("Загрузка конфигурации модуля завершена.")

	query := "SELECT data, md5(data::text), update_time FROM config WHERE is_active='t';"
	err = db.QueryRow(query).Scan(&data, &CfgMD5Sum, &updCfgTime)
	if err != nil {
		l.Warn(17901, err.Error())
		return updCfgTime, err
	}

	schemaLoader, documentLoader := loaders(data)

	if err = Check("Конфигурация из СУБД", schemaLoader, documentLoader); err != nil {
		return updCfgTime, err
	}

	err = json.Unmarshal([]byte(data.String), &conf)
	if err != nil {
		l.Err(10901, err.Error())
		return updCfgTime, err
	}

	return updCfgTime, nil
}

// Check - функция проверки конфигурационного файла программы
func Check(s string, ls gojsonschema.JSONLoader, ld gojsonschema.JSONLoader) error {
	l.Debug("Проверка конфигурации модуля по схеме...")
	defer l.Debug("Проверка конфигурации модуля по схеме завершена.")

	var (
		result *gojsonschema.Result
		err    error
	)

	if result, err = gojsonschema.Validate(ls, ld); err != nil {
		l.Err(10902, ":"+s+"- не удалось выполнить проверку по схеме,"+err.Error())
		return err
	}

	if !result.Valid() {
		var errs []string
		l.Err(10903, s, " - проверка по схеме вернула ошибки.")
		for _, desc := range result.Errors() {
			l.Err(10904, desc)
			errs = append(errs, desc.String())
		}

		return errors.New("Проверка по схеме вернула ошибки :: " + strings.Join(errs, " , "))
	}

	l.Debug(s, " - проверка успешна. ошибок не найдено.")
	return nil
}

// GetHandler - функция выгрузки конфига
func GetHandler(conf *Config, db *sql.DB) ([]byte, error) {

	l.SetPrefix(conf.JournalName)
	l.SetLogLevel(conf.LogLevel)
	l.Debug("Вызов config.GetHandler")
	defer l.Debug("Завершена обработка config.GetHandler")

	type MainPartOfConfig struct {
		Module      interface{} `json:"module,omitempty"`
		Functional  interface{} `json:"functional,omitempty"`
		Environment interface{} `json:"environment,omitempty"`
		Integration interface{} `json:"integration,omitempty"`
	}

	var (
		res MainPartOfConfig
	)

	res = MainPartOfConfig{Module: conf.Module, Functional: conf.Functional, Environment: conf.Environment, Integration: conf.Integration}
	r, _ := json.Marshal(res)

	return r, nil
}

// SaveHandler - функция сохранения конфига
func SaveHandler(r io.Reader, conf *Config, db *sql.DB) error {

	l.SetPrefix(conf.JournalName)
	l.SetLogLevel(conf.LogLevel)
	l.Debug("Вызов config.SaveHandler")
	defer l.Debug("Завершена обработка config.SaveHandler")

	var (
		err error
	)

	err = db.Ping()
	if err != nil {
		l.Warn(17904, err.Error())
		return err
	}

	if r == nil {
		l.Warn(14901)
		return err
	}

	body, err := ioutil.ReadAll(r)
	if err != nil {
		l.Warn(14902, err.Error())
		return err
	}

	l.Debug("Получена конфигурация для сохранения: ", string(body))

	inData := &configReq{}
	err = json.Unmarshal(body, inData)
	if err != nil {
		l.Warn(10908, err.Error())
		return err
	}

	schemaLoader := gojsonschema.NewStringLoader(SQLConfigSchema)
	documentLoader := gojsonschema.NewBytesLoader(body)

	if err = Check("Новая конфигурация", schemaLoader, documentLoader); err != nil {
		l.Warn(10907)
		return err
	}

	// Добавляем новую запись в конфиг с именем - текущие дата+время в секундах с точкой
	query := "INSERT into config (name,data) VALUES (EXTRACT(EPOCH FROM now())::text, $1)"
	_, err = db.Exec(query, string(body))
	if err != nil {
		l.Err(17905, err.Error())
		return err
	}

	return nil
}

// SchemaHandler - функция выгрузки JSON схемы для SQL конфига
func SchemaHandler() ([]byte, error) {
	var (
		err error
		res schemaResp
	)

	res.Result = json.RawMessage(SQLConfigSchema)
	schema, err := json.Marshal(res.Result)
	if err != nil {
		l.Warn(10909)
		return nil, err
	}

	return schema, nil
}
