FROM golang:1.17

WORKDIR ~/go/src/getherscan
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...