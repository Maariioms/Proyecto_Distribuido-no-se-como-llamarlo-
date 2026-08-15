FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /soc-api ./cmd/api

FROM alpine:3.19
WORKDIR /app
COPY --from=build /soc-api /app/soc-api
EXPOSE 8080
CMD ["/app/soc-api"]
