FROM golang:1.22-alpine AS build
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o /status-api ./cmd/api

FROM alpine:3.19
WORKDIR /app
COPY --from=build /status-api /app/status-api
EXPOSE 8080
CMD ["/app/status-api"]
