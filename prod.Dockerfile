FROM golang:1.24.1-alpine AS base

WORKDIR /app

# ===========================
FROM base AS build

COPY --link go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o go-sso.app cmd/sso/main.go

# ===========================
FROM base

COPY --from=build /app/go-sso.app /app/go-sso.app

CMD ["./go-sso.app"]
