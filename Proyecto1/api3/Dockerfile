# Build etapa: compila el binario de Go.
# Usamos golang:1.23-alpine para tener las herramientas de Go.
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

# Runtime etapa: solo copiamos el binario compilado.
# Alpine es una imagen minima (~5MB), no necesita Go instalado.
# Esta imagen se sube a Zot y se corre en VM2 con podman.
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
