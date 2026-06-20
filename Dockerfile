FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE
RUN case "$SERVICE" in \
      opsweaver-server|opsweaver-worker|opsweaver-gateway) ;; \
      *) echo "Invalid SERVICE: $SERVICE" >&2; exit 1 ;; \
    esac && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/service "./cmd/${SERVICE}"

FROM alpine:3.22

RUN apk add --no-cache ca-certificates && \
    addgroup -S opsweaver && adduser -S -G opsweaver opsweaver

WORKDIR /app

COPY --from=build /out/service /app/service
COPY deploy/config/compose.yaml /app/config.yaml

USER opsweaver

ENTRYPOINT ["/app/service"]
CMD ["-config", "/app/config.yaml"]
