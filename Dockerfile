# ==============================================================================

FROM node:alpine3.16 AS frontend_builder

WORKDIR /frontend
COPY ./frontend/package.json /frontend/package.json

ENV NODE_OPTIONS=--openssl-legacy-provider
RUN yarn --ignore-engines

COPY ./frontend/preact.config.js /frontend/preact.config.js
COPY ./frontend/prerender-urls.json /frontend/prerender-urls.json
COPY ./frontend/tsconfig.json /frontend/tsconfig.json
COPY ./frontend/src /frontend/src
RUN npm run build

# ==============================================================================

FROM golang:1.26-alpine AS backend_builder
ARG TARGETOS=linux
ARG TARGETARCH=amd64
WORKDIR /backend
RUN apk add --no-cache ca-certificates
COPY ./backend/go.mod ./backend/go.sum ./
RUN go mod download
COPY ./backend .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o main

# ==============================================================================

FROM scratch
COPY --from=backend_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=backend_builder /backend/main /app/server
COPY --from=frontend_builder /frontend/build /app/static
COPY ./config.json /app/config.json
USER 65532:65532
ENTRYPOINT ["/app/server"]
