# TODO: multi-stage build for kokoroya-backend
#
# FROM golang:1.26-alpine AS builder
# WORKDIR /app
# COPY go.mod go.sum ./
# RUN go mod download
# COPY . .
# RUN go build -o /app/bin/api ./cmd/api
#
# FROM alpine:3.20
# COPY --from=builder /app/bin/api /usr/local/bin/api
# COPY config/config.json /config/config.json
# EXPOSE 8080
# ENTRYPOINT ["api"]
