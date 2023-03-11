FROM golang:1.20.0-alpine3.17

WORKDIR /app

COPY . ./
RUN go mod download

RUN go build cmd/server/main.go

EXPOSE 1323

CMD [ "./main" ]
