FROM golang:latest

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

COPY telegram-files ./

RUN go mod download

COPY . .

RUN go build -o watch .

CMD ["./watch"]