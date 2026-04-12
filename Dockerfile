FROM golang:1.24 AS build
WORKDIR /src

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates git \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

ENV GOPROXY=direct
ENV GOSUMDB=off

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/worker ./cmd/worker

FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates wget procps \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/worker /usr/local/bin/worker
COPY --from=build /src/migrations ./migrations

ENV HTTP_ADDR=:8081

CMD ["api"]
