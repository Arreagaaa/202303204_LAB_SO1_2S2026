FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY go.mod .
COPY proto ./proto
COPY consumer ./consumer
WORKDIR /src/consumer
RUN go build -o /out/consumer .

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/consumer /consumer
ENTRYPOINT ["/consumer"]
