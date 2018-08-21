package jsgen

// Generate - создает JSON-схему из конфига
func (s *TSchema) GenerateSchema(config map[string]interface{}) error {
	recurseConfig(s, "", config)

	return nil
}

func recurseConfig(s *TSchema, prevID string, node interface{}) {
	typ := classify(node)

	if typ == "object" {
		for k, v := range node.(map[string]interface{}) {
			id := prevID + "/properties/" + k
			typ := classify(v)

			ns := s.AddProperty(k, id, typ)
			s.AddRequired(k)

			recurseConfig(ns, id, v)
		}
	}

	return
}

func classify(node interface{}) string {
	var ret string

	switch node.(type) {
	case map[string]interface{}:
		ret = "object"
	case int, int32, int64, float32, float64:
		t := node.(float64)
		it := int(t)

		if float64(it) == t {
			ret = "integer"
		} else {
			ret = "number"
		}
	case string:
		ret = "string"
	case bool:
		ret = "boolean"
	}

	return ret
}
