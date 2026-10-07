FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY . .
COPY ../go.mod ../go.mod
COPY ../proto ../proto
RUN go build -o /out/go-ingest ./go-ingest

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/go-ingest /go-ingest
EXPOSE 8081
ENTRYPOINT ["/go-ingest"]
