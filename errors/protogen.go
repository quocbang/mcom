package errors

// NOTE: $GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.11.3 should
//       match the version in go.mod and Dockerfile. Once everyone is on non-Windows
//       systems we can use go list to get the path

//go:generate protoc --go_out=paths=source_relative:. code.proto
