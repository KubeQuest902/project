FROM alpine:3.12

RUN GRPC_HEALTH_PROBE_VERSION=v0.4.24 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.37/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

WORKDIR /app

EXPOSE 8080

COPY ./app /app/app
COPY ./web /app/web

RUN chmod -R 755 /app/web && \
    chown -R root:root /app/web

RUN chmod +x /app/app

ENTRYPOINT ["/app/app"]