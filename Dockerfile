FROM node:24.15-alpine AS frontend
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.27.3-alpine AS backend
WORKDIR /src
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/precizzle .

FROM alpine:3.23
RUN adduser -D -H -u 10001 precizzle
COPY --from=backend /out/precizzle /usr/local/bin/precizzle
COPY --from=frontend /src/dist /srv/precizzle
# The container listens on every interface so the published port reaches it;
# the documented run command publishes it only on the host loopback (D89).
ENV PRECIZZLE_ADDR=:8080 \
    PRECIZZLE_STATIC_DIR=/srv/precizzle
EXPOSE 8080
USER precizzle
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --start-interval=1s \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["precizzle"]
