FROM golang:1.21-alpine
WORKDIR /app
COPY main.go .
RUN go build -o echo-go main.go
EXPOSE 8081
CMD ["./echo-go"]
