package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"parti/jsgen/jsgen"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

const (
	version = "1.0"
)

func main() {
	var (
		err    error
		config map[string]interface{}
		schema *jsgen.TSchema
		queues []string
	)

	flagConfigFile := flag.String("c", "../cfg/config_from_db.json", "файл с конфигом из базы")
	flagSchemaFile := flag.String("s", "../config/sql_schema.json", "Файл с JSON-схемой")
	flagNoLearn := flag.Bool("n", false, "Не брать поля title из существующей схемы")
	flagQueuesList := flag.String("q", "mail,err", "Перечень наименований очередей (через запятую)")
	flagUsage := flag.Bool("h", false, "Помощь")

	flag.Parse()

	if *flagUsage {
		flag.Usage()
		os.Exit(0)
	}

	// if ok, e := validateConfig(*flagConfigFile, *flagSchemaFile); !ok {
	// 	fmt.Printf("Конфигурация не прошла валидацию. Необходимо переделать схему. Список ошибок:\n\n%s\n", e)
	// 	return
	// }

	learnMap := make(map[string]string)

	if schema, err = getSchema(*flagSchemaFile); err != nil {
		log.Println("Ошибка обработки файла схемы, работаем без схемы. Текст ошибки: ", err.Error())
	} else {
		if !*flagNoLearn {
			jsgen.Learn(schema, learnMap)
		}
	}

	if queues = getQueues(*flagQueuesList); len(queues) > 0 {
		jsgen.AppendQueues(queues)
	}

	if config, err = getConfig(*flagConfigFile); err != nil {
		log.Fatal("Ошибка обработки конфига", err.Error())
	}

	// создаем объект схемы
	s := jsgen.New()

	// устанавливаем top level параметры
	s.SetRoot()
	// генерируем схему по конфигу
	s.GenerateSchema(config)
	// устанавливаем заголовки
	s.SetTitles(learnMap)

	fmt.Println(s)
}

// getConfig - получение конфига из файла
func getConfig(filename string) (map[string]interface{}, error) {
	var (
		ret map[string]interface{}
		j   []byte
		err error
	)

	if j, err = ioutil.ReadFile(filename); err != nil {
		return nil, err
	}

	if err = jsgen.Decode(j, &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

// getSchema - получение схемы из файла
func getSchema(filename string) (*jsgen.TSchema, error) {
	var (
		j   []byte
		err error
	)

	ret := jsgen.New()

	if j, err = ioutil.ReadFile(filename); err != nil {
		return nil, err
	}

	ret.Load(j)

	return ret, nil
}

// getQueues - получение списка очередей
func getQueues(q string) []string {
	return strings.Split(q, ",")
}

// validateConfig - валидация конфига через схему. Временно непонятно, зачем
func validateConfig(flagConfigFile string, flagSchemaFile string) (bool, string) {
	var (
		ret string
	)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	schemaLoader := gojsonschema.NewReferenceLoader("file://" + dir + "/" + flagSchemaFile)
	documentLoader := gojsonschema.NewReferenceLoader("file://" + dir + "/" + flagConfigFile)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		panic(err.Error())
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			ret = ret + fmt.Sprintf("- %s\n", desc)
		}
	}

	return result.Valid(), ret
}
