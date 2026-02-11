# Caddy custom build for Hearth
# Required modules: caddy-l4 (Layer 4 TLS SNI routing) + caddy-yaml (YAML config adapter)
# Runtime memory: ~15-20MB — fits in OS headroom budget

FROM caddy:2-builder AS builder

RUN xcaddy build \
    --with github.com/abiosoft/caddy-yaml \
    --with github.com/mholt/caddy-l4

FROM caddy:2

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
