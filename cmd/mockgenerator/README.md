# About Jennifer mockgenerator

## Features

Automatically generate `mock/mock_func.go` according to the method signatures in `dm.go`

## Usage

```powershell
cd .\cmd\mockgenerator\
go run main.go
```

or run `go generate` in `./generator.go`

## Notice

The method signatures in `dm.go` must comply with the following specifications.

* The number of arguments/return values

  |          | Arguments | Return Value |
  | -------- | --------- | ------------ |
  | &#9745;  | 1         | 2            |
  | &#9745;  | 2         | 1            |
  | &#9745;  | 2         | 2            |
  | &#x2612; | 3         | 1            |
  | &#9745;  | 3         | 2            |
  | &#x2612; | others    | others       |

* the first argument must be `context.Context`

* the last return value must be `error`

The methods which **cannot** satisfy the specifications above should be written in `mock/mock.go` as exceptions.

## Reference Packages

* [reflect](https://pkg.go.dev/reflect)
* [jennifer](https://pkg.go.dev/github.com/dave/jennifer@v1.4.1/jen)
