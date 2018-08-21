package jsgen

import (
	"strings"
)

// Learn - собираем title в имеющейся схеме
func Learn(s *TSchema, ret map[string]string) {
	if s.Title != "" && !strings.HasPrefix(s.Title, "The ") {
		ret[s.ID] = s.Title
	}

	for _, v := range s.Properties {
		Learn(v, ret)
	}
}
