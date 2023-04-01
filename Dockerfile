# Alpine is chosen for its small footprint
# compared to Ubuntu
FROM golang:1.20.2-alpine3.17

WORKDIR /app

# Download necessary Go modules
COPY go.mod ./
RUN go mod download

COPY *.go ./
COPY .env ./
RUN go mod tidy

RUN go build -o main .

EXPOSE ${PORT}

CMD ["./main"]