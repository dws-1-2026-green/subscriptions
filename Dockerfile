ARG GO_VERSION=1.26
FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGET
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${TARGET}


FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=build /out/app /app

USER nonroot:nonroot
ENTRYPOINT ["/app"]