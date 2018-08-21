package alxlogger

import (
	"encoding/json"
	"fmt"
)

// DebugJSON - вывод JSON для отладки в stdout
func DebugJSON(title string, data interface{}) {
	j, _ := json.MarshalIndent(data, "\t", "  ")
	fmt.Printf("%s:\n%s\n", title, j)
}
