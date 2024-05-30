from golang:1.22-alpine3.20 as build

WORKDIR /app

COPY . .
RUN go mod download
RUN go build -o /go/tls-impersonate


from golang:1.22-alpine3.20

RUN mkdir /go/app

ENV GOPATH=/go
WORKDIR /go/app

COPY --from=build /go/tls-impersonate /go/app
