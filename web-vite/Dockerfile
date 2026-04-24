# syntax=docker/dockerfile:1.7

# Stage 1 — build
FROM node:22-alpine AS build
WORKDIR /app
RUN corepack enable
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

# Precompress static assets — nginx serves via brotli_static / gzip_static.
# Build-time cost pays off forever at request time.
RUN apk add --no-cache brotli \
 && find dist -type f \( -name '*.js' -o -name '*.css' -o -name '*.html' -o -name '*.svg' -o -name '*.json' \) \
      -exec gzip -k -9 {} \; \
      -exec brotli -k -q 11 {} \;

# Stage 2 — runtime (Chainguard distroless nginx)
FROM cgr.dev/chainguard/nginx:latest
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY public/config.js.template /etc/nginx/templates/config.js.template
COPY docker-entrypoint.d/ /docker-entrypoint.d/
# distroless nginx runs envsubst on /etc/nginx/templates/*.template at
# boot via /docker-entrypoint.d/ scripts — our 10-envsubst-config.sh
# writes the final /config.js into the html root.
