package hospital

import (
	"fmt"
	"strings"
)

// Person is embedded by value. Identity is immutable after construction.
type Person struct {
	id   string
	name string
	age  int
}

func newPerson(id, name string, age int) (Person, error) {
	id, name = strings.TrimSpace(id), strings.TrimSpace(name)
	if id == "" || name == "" || age < 0 || age > 130 {
		return Person{}, fmt.Errorf("identidad inválida: ID y nombre obligatorios; edad entre 0 y 130")
	}
	return Person{id: id, name: name, age: age}, nil
}

func (p Person) ID() string   { return p.id }
func (p Person) Name() string { return p.name }
func (p Person) Age() int     { return p.age }
