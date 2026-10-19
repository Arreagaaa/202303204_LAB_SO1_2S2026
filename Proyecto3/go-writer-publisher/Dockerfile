FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY go-writer-publisher ./go-writer-publisher
WORKDIR /src/go-writer-publisher
RUN go build -o /out/go-writer-publisher .
FROM gcr.io/distroless/base-debian12
COPY --from=build /out/go-writer-publisher /go-writer-publisher
EXPOSE 8083
ENTRYPOINT ["/go-writer-publisher"]
