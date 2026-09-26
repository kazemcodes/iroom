# syntax=docker/dockerfile:1
# Single image that serves both the API and the built SvelteKit frontend.
# Works for docker-compose (behind Caddy) and for single-port container
# platforms like Hugging Face Spaces, Render and Fly.io.

FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /server .
COPY --from=frontend /app/web/.svelte-kit/output/client ./static
COPY config.yaml .
# uploads/recordings are scratch space; with an external DB these are the only
# ephemeral bits left, and they are intentionally not persisted.
RUN mkdir -p uploads recordings data

# Must be 0.0.0.0 for the platform proxy to reach us, and the port comes from
# $PORT (Hugging Face Spaces, Render, Fly.io) or SERVER_PORT (local/compose).
ENV SERVER_HOST=0.0.0.0 \
    PORT=7860 \
    UPLOAD_DIR=uploads

EXPOSE 7860
EXPOSE 8080

CMD ["./server"]
