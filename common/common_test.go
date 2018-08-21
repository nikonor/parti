package common

import (
	"testing"
)

func TestObjectIDGen(t *testing.T) {
	q := ObjectIDGen()
	println(q)
}
