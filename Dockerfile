FROM golang:1.26-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/subscriptions-worker ./cmd/worker


FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=build /out/subscriptions-worker /subscriptions-worker

USER nonroot:nonroot
ENTRYPOINT ["/subscriptions-worker"]