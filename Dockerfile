FROM golang:1.22-alpine
RUN apk add --no-cache git
RUN git clone https://github.com/luiz734/chatapp-api-workers

WORKDIR chatapp-api-workers
RUN rm go.sum go.mod
RUN go mod init chatapp-api-workers
RUN go mod tidy

CMD go run .
