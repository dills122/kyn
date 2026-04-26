FROM golang:1.23 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build -trimpath -ldflags="-s -w" -o /out/kyn ./cmd/kyn

FROM gcr.io/distroless/static:nonroot

COPY --from=build /out/kyn /usr/local/bin/kyn

ENTRYPOINT ["/usr/local/bin/kyn"]
