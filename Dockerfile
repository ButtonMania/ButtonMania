# ==============================================================================

FROM node:alpine3.16 as frontend_builder

WORKDIR /frontend
COPY ./frontend/package.json /frontend/package.json

ENV NODE_OPTIONS --openssl-legacy-provider
RUN yarn

COPY ./frontend/preact.config.js /frontend/preact.config.js
COPY ./frontend/prerender-urls.json /frontend/prerender-urls.json
COPY ./frontend/tsconfig.json /frontend/tsconfig.json
COPY ./frontend/src /frontend/src
RUN npm run build

# ==============================================================================

FROM golang AS backend_builder
WORKDIR /backend
COPY ./backend .
RUN go mod download
RUN update-ca-certificates
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

# ==============================================================================

FROM scratch
COPY --from=backend_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=backend_builder /backend/main /app/server
COPY --from=frontend_builder /frontend/build /app/static
ENTRYPOINT ["/app/server"]
