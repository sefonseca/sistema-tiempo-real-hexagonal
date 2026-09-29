# ---------- Etapa 1: compilar un binario estático ----------
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---------- Etapa 2: imagen mínima de ejecución ----------
FROM alpine:3.20
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/server /usr/local/bin/server
USER app
ENV PORT=8080
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- "http://127.0.0.1:${PORT}/healthz" || exit 1
ENTRYPOINT ["/usr/local/bin/server"]
