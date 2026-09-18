FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /events-api ./cmd/api

FROM alpine:3.19
WORKDIR /app
COPY --from=build /events-api /app/events-api
EXPOSE 8080
CMD ["/app/events-api"]
