FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o round-robin-api ./cmd/main.go
EXPOSE 8080
CMD ["./round-robin-api"]
