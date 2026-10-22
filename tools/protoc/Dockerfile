# imagen solo pa generar codigo de los .proto
# asi no instalas protoc en la compu
FROM golang:1.22-alpine

# protoc que es el compilador de .proto, mas los includes pa google/protobuf/timestamp.proto
RUN apk add --no-cache protobuf protobuf-dev

# los dos plugins de go
# uno saca los mensajes .pb.go y el otro el grpc _grpc.pb.go
# versiones fijas pa que sirvan con go 1.22
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2 \
 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0

ENV PATH="/go/bin:${PATH}"
