FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY go.mod .
COPY proto ./proto
COPY go-ingest-client ./go-ingest-client
WORKDIR /src/go-ingest-client
RUN go build -o /out/go-ingest-client .
FROM gcr.io/distroless/base-debian12
COPY --from=build /out/go-ingest-client /go-ingest-client
EXPOSE 8082
ENTRYPOINT ["/go-ingest-client"]
