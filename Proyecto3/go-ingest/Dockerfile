FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY go.mod .
COPY proto ./proto
COPY go-ingest ./go-ingest
WORKDIR /src/go-ingest
RUN go build -o /out/go-ingest .

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/go-ingest /go-ingest
EXPOSE 8081
ENTRYPOINT ["/go-ingest"]
