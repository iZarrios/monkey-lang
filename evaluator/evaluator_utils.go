package evaluator

import "github.com/iZarrios/monkey-lang/object"

// Instead of allocating everytime we stumble upon a distinct value, we are just going to use the same variables
var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	} else {
		return FALSE
	}
}
