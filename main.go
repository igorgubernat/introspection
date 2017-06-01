package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"reflect"
	"time"
)

type (
	A struct {
		E int       `json:"e"`
		F bool      `json:"f" default:"false" description:"Some type F"`
		G time.Time `json:"g"`
	}

	B struct {
		C string     `json:"c"`
		D gocql.UUID `json:"d"`
		Z A          `json:"z"`
		S []string   `json:"s" default:"['a', 'b']" description:"A slice of string"`
		T []A        `json:"t" description:"A slice of structures"`
	}
)

// Field contains info about a structure field
type Field struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
}

// Node in the tree structure of types
type Node struct {
	f        Field
	parent   *Node
	children []*Node
}

// Map with string representations of some types
var TypeAliaces = map[reflect.Type]string{
	reflect.TypeOf(gocql.UUID{}): "uuid",
	reflect.TypeOf(time.Time{}):  "timestamp",
}

// End result, flattened
var fields []Field

func main() {
	b := B{}
	fmt.Println(GetMeta(b))
}

// GetMeta takes a data structure and returns json representation of it's structure
func GetMeta(i interface{}) string {
	v := reflect.Indirect(reflect.ValueOf(i))
	root := Node{}
	root.parent = nil
	root.children = make([]*Node, 0)
	getMeta(&root, v)
	flatten(&root)

	b, err := json.Marshal(fields)
	if err != nil {
		fmt.Printf("Error encoding json: %v\n", err)
	}
	fields = fields[:0]
	return string(b)
}

// getMeta is a recursive function that builds the tree representation of evaluated data type
func getMeta(n *Node, v reflect.Value) {
	if v.Type() == reflect.TypeOf(time.Time{}) {
		return
	}
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)

			f := Field{}

			var ok bool
			f.Type, ok = TypeAliaces[field.Type]
			if !ok {
				f.Type = field.Type.String()
			}

			f.Description = field.Tag.Get("description")
			f.Default = field.Tag.Get("default")

			json := field.Tag.Get("json")
			if json != "" {
				f.Name = json
			} else {
				f.Name = field.Name
			}
			node := Node{}
			node.parent = n
			node.f = f
			node.children = make([]*Node, 0)
			n.children = append(n.children, &node)
			getMeta(&node, v.Field(i))
		}
	} else if v.Kind() == reflect.Slice {
		if n.f.Name != "" {
			n.f.Name = "[]" + n.f.Name
		} else {
			n.f.Name = v.Type().String()
		}
		sliceVal := reflect.MakeSlice(v.Type(), 1, 1)
		getMeta(n, sliceVal.Index(0))
	}
}

// flatten transforms the tree created by getMeta into slice
func flatten(node *Node) {
	if len(node.children) == 0 {
		fields = append(fields, node.f)
		return
	}
	for _, child := range node.children {
		if node.f.Name != "" {
			child.f.Name = node.f.Name + "." + child.f.Name
		}
		if node.f.Description != "" && child.f.Description != "" {
			child.f.Description = node.f.Description + ". " + child.f.Description
		}
		flatten(child)
	}
}
