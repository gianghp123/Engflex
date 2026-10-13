FROM golang:1.26.4-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api ./
# Swagger docs are not committed; generate them so /swagger works in the image.
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6 && \
    "$(go env GOPATH)/bin/swag" init -g cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
# Coolify's Dockerfile-buildpack healthcheck execs curl (or wget) inside the
# container — without it every deploy is marked unhealthy and rolled back.
RUN apk add --no-cache ca-certificates tzdata curl
COPY --from=build /out/server /server
EXPOSE 8000
ENTRYPOINT ["/server"]
