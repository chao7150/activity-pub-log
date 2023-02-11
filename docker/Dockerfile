FROM golang:1.20.0-alpine3.17

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download


COPY server.go ./

RUN go build server.go

EXPOSE 1323

CMD [ "./server" ]
