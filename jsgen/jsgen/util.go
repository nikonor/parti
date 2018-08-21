package jsgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// DecodeWrapper - обертка над json.Decoder.Decode
func DecodeWrapper(dec *json.Decoder, out interface{}) error {
	if err := dec.Decode(&out); err != nil {
		switch err.(type) {
		case *json.UnmarshalTypeError:
			return fmt.Errorf("Параметр " + err.(*json.UnmarshalTypeError).Field + " имеет неверный формат")
		default:
			return fmt.Errorf("Тело запроса содержит некорректную структуру JSON")
		}
	}

	return nil
}

// DecodeBytes - декодирование []bytes с проверками
func DecodeBytes(in []byte, out interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(in))
	return DecodeWrapper(dec, out)
}

// DecodeReader - декодирование io.Reader с проверками
func DecodeReader(in io.Reader, out interface{}) error {
	dec := json.NewDecoder(in)
	return DecodeWrapper(dec, out)
}

// Decode - функция декодирования полученного запроса
//      in м.б. либо []byte, либо io.Reader
func Decode(in interface{}, out interface{}) error {
	switch in.(type) {
	case []byte:
		return DecodeBytes(in.([]byte), out)
	case io.Reader:
		return DecodeReader(in.(io.Reader), out)
	}
	return fmt.Errorf("Неизвестный формат входящих данных")
}

// AppendQueues - добавление нужного количества мапов очередей с правильными ключами
func AppendQueues(queues []string) {
	for k, v := range queuesHeadMap {
		titlesMap[k] = v
	}

	for _, q := range queues {
		for k, v := range queuesMap {
			titlesMap[fmt.Sprintf(k, q)] = v
		}
	}
}

// Lookup - поиск русского соответствия в мапах
func Lookup(what string, learnMap map[string]string) string {
	var (
		ret string
		ok  bool
	)

	if ret, ok = learnMap[what]; ok {
		return ret
	}

	if ret, ok = titlesMap[what]; ok {
		return ret
	}

	if ret, ok = customTitlesMap[what]; ok {
		return ret
	}

	for k, v := range regexpTitlesMap {
		if match, _ := regexp.MatchString(k, what); match {
			return v
		}
	}

	a := strings.Split(what, "/")

	if len(a) > 0 {
		ret = a[len(a)-1]
	}

	return ret
}
