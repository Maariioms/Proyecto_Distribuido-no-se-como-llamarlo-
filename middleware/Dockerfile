FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /soc-middleware .

FROM alpine:3.19
WORKDIR /app
COPY --from=build /soc-middleware /app/soc-middleware
EXPOSE 9000
CMD ["/app/soc-middleware"]
