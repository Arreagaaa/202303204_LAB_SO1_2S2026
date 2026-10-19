FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY go.mod .
COPY metrics ./metrics
RUN go build -o /out/metrics ./metrics
FROM gcr.io/distroless/base-debian12
COPY --from=build /out/metrics /metrics
EXPOSE 9091
ENTRYPOINT ["/metrics"]
