# ── Build stage ───────────────────────────────────────────
FROM golang:1.25.0 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /chat-app-backend

# ── Runtime stage ─────────────────────────────────────────
FROM alpine:3.19 AS runtime
RUN adduser -D -u 1000 nonroot

COPY --from=build /chat-app-backend /chat-app-backend

USER nonroot
EXPOSE 8080
CMD ["/chat-app-backend"]