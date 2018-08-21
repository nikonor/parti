package jsgen

import (
	"encoding/json"
	"sort"
	"strings"
)

// New - новый объект схемы
func New() *TSchema {
	return &TSchema{}
}

// SetRoot - установка корневых параметров схемы
func (s *TSchema) SetRoot() {
	s.ID = "http://example.com/example"
	s.Type = "object"
	s.Schema = "http://json-schema.org/draft-07/schema#"
	s.Definitions = &TSchemaDefinitions{}
	s.Properties = TSchemaProperties{}
}

// Load - загрузка схемы
func (s *TSchema) Load(in interface{}) error {
	return s.Decode(in)
}

// Decode - функция декодирования полученного запроса
//      in м.б. либо []byte, либо io.Reader
func (s *TSchema) Decode(in interface{}) error {
	return Decode(in, s)
}

// SetTitle - Установка title у элемента
func (s *TSchema) SetTitle(title string) {
	if s.Title == "" || strings.HasPrefix(s.Title, "The ") {
		s.Title = title
	}
}

// SetTitles - рекурсивная установка title
func (s *TSchema) SetTitles(learnMap map[string]string) error {
	for _, v := range s.Properties {
		if title := Lookup(v.ID, learnMap); title != "" {
			v.SetTitle(title)
		}
		v.SetTitles(learnMap)
	}

	return nil
}

// AddProperty - добавление новой подсхемы
func (s *TSchema) AddProperty(k string, id string, typ string) *TSchema {
	newProperty := &TSchema{
		ID:         id,
		Type:       typ,
		Properties: TSchemaProperties{},
	}

	s.Properties[k] = newProperty

	return newProperty
}

// AddRequired - добавление элемента в массив required
func (s *TSchema) AddRequired(elem string) {
	// s.Required = append(s.Required, elem)
	s.Required = SortedInsert(s.Required, elem)
}

// String - стрингер для проперти
func (p TSchemaProperties) String() string {
	j, _ := json.MarshalIndent(p, "  ", "  ")
	return string(j)
}

// String - стрингер для схемы
func (s *TSchema) String() string {
	j, _ := json.MarshalIndent(s, "  ", "  ")
	return string(j)
}

func SortedInsert(data []string, el string) []string {
	index := sort.Search(len(data), func(i int) bool { return data[i] > el })
	data = append(data, "")
	copy(data[index+1:], data[index:])
	data[index] = el
	return data
}
