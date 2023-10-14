# Convert

Generates structures converting code

## Installation

```shell
go install github.com/sabahtalateh/convert@latest
```

You may also need to modify your `~/.[bash|zsh|fish..]rc` with

```shell
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```

So modifications to take effect run
```shell
source ~/.[bash|zsh|fish..]rc
```

## Usage

### Both `In` & `Out` structures exist

```go
// Suppose there are 2 structs

// In
type In struct {
	Bool       bool
	String     string
	Int        int
	Bytes      []byte
	SomeStruct *SomeStruct
}

// and Out
type Out struct {
	Bool       bool
	String     string
	Int        int
	Bytes      []byte
	SomeStruct *SomeStruct
}

//go:generate convert
func Convert(in In) Out {
	// Function body may be emtpy or not. It will be replaced
}
```

Then run

```shell
go generate ./...
```

And see new body of `Convert` function

```go
func Convert(in In) Out {
	var out Out
	
	out.Bool = in.Bool
	out.String = in.String
	out.Int = in.Int
	out.Bytes = in.Bytes
	out.SomeStruct = in.SomeStruct
	
	return out
}
```

What changed:

1. Function body replaced
2. `//go:generate convert` removed. `convert` designed to be a one-time command that will generate boilerplate code that will further be modified by user, so it removed after usage

### Only `In` structure exists

```go
// Suppose there is only In structure

// In
type In struct {
	Bool       bool
	String     string
	Int        int
	Bytes      []byte
	SomeStruct *SomeStruct
}

//go:generate convert
func Convert(in In) Out {}
```

Then after
```shell
go generate ./...
```

`Out` will be created with all the same exported fields as `In` have

```go
type Out struct {
	Bool       bool
	String     string
	Int        int
	Bytes      []byte
	SomeStruct *SomeStruct
}

func Convert(in In) Out {
	var out Out

	out.Bool = in.Bool
	out.String = in.String
	out.Int = in.Int
	out.Bytes = in.Bytes
	out.SomeStruct = in.SomeStruct

	return out
}
```

## Slices

Conversion between slices also allowed

```go
type In struct {
	SomeStruct *SomeStruct
}

type Out struct {
	SomeStruct *SomeStruct
}

type SomeStruct struct {}

//go:generate convert
//func Convert(in []*In) []Out {}
func Convert(in []*In) []*Out {
	var out []*Out

	for _, x := range in {
		var out2 &Out
		out2.SomeStruct = x.SomeStruct

		out = append(out, out2)
	}

	return out
}
```

## Auto `&` and `*`

Simple `&` and `*` operations automatically performed on `In` and `Out` itself and on theirs fields


```go
// valid
func Convert(in *In) Out {}

// valid
func Convert(in In) *Out {}
```

```go
// valid
type In struct {
	S string
}

type Out struct {
	S *string
}

func Convert(in *In) Out {
	var out Out
	out.S = &in.S // <-- auto &
	return out
}
```

```go
// valid
type In struct {
	S *string
}

type Out struct {
	S string
}

func Convert(in *In) Out {
	var out Out
	out.S = *in.S // <-- auto *
	return out
}
```
