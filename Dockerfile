FROM golang:1.17

RUN     mkdir /app
WORKDIR /app
ADD     go.mod main.go /app/
RUN     go build

CMD     ./udp-echo

