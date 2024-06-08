FROM golang:1.22.3

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN go test ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

EXPOSE 8080

CMD ["./main"]
