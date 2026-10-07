FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY . .
COPY ../go.mod ../go.mod
COPY ../proto ../proto
RUN go build -o /out/go-writer ./go-writer

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/go-writer /go-writer
EXPOSE 50051
ENTRYPOINT ["/go-writer"]
