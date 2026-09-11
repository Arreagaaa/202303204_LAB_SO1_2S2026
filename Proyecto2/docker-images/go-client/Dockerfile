FROM golang:1.23-alpine AS build
WORKDIR /src
COPY main.go .
RUN go build -o /go-client main.go

FROM alpine:3.20
COPY --from=build /go-client /usr/local/bin/go-client
CMD ["/usr/local/bin/go-client"]
