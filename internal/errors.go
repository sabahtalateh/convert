package internal

import (
	"errors"
)

var (
	errUnsupportedType    = errors.New("unsupported type. supported: MyType, *MyType, []MyType, []*MyType")
	errUnsupportedTypeRef = errors.New("unsupported type")
	errBasicType          = errors.New("basic type not supported")
	errTypeNotFound       = errors.New("type not found")
	errNotStruct          = errors.New("not struct")
)
