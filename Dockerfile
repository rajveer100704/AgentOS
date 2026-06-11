FROM --platform=$BUILDPLATFORM golang:1.26.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} go build -o agentos ./cmd/agentos
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} go build -o agentctl ./cmd/agentctl

FROM alpine:3.24
RUN apk add --no-cache ca-certificates curl jq bash
WORKDIR /app
COPY --from=builder /app/agentos .
COPY --from=builder /app/agentctl /usr/local/bin/
COPY configs/demo.yaml /app/configs/agentos.yaml
COPY configs/policy-packs/ /app/configs/policy-packs/
COPY scripts/demos/demo.sh /app/scripts/
RUN chmod +x /app/scripts/*.sh
EXPOSE 8080 8081 8082
ENTRYPOINT ["./agentos"]
