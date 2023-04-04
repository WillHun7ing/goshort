FROM golang:1.20

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /goshort
EXPOSE ${PORT}

CMD ["/goshort"]