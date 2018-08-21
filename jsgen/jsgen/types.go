package jsgen

// TSchema -- JSON-схема
type TSchema struct {
	Definitions *TSchemaDefinitions `json:"definitions,omitempty"`
	Schema      string              `json:"$schema,omitempty"`
	ID          string              `json:"$id" validate:"required"`
	Type        string              `json:"type" validate:"required"`
	Title       string              `json:"title"`
	Properties  TSchemaProperties   `json:"properties,omitempty"`
	Required    []string            `json:"required,omitempty"`
}

// TSchemaDefinitions -- тип для определений схемы
type TSchemaDefinitions struct{}

// TSchemaPropertyType -- алиас string
type TSchemaPropertyType string

// TSchemaProperties -- элемент схемы
type TSchemaProperties map[string]*TSchema
