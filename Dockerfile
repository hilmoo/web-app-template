# syntax=docker/dockerfile:1

ARG NODE_VERSION=24
ARG GO_VERSION=1.27
ARG PNPM_VERSION=12.13.0

# ---- Frontend build: produces frontend/dist ----
FROM node:${NODE_VERSION}-slim AS frontend
ARG PNPM_VERSION
ENV PNPM_HOME=/pnpm PATH=/pnpm:$PATH
RUN corepack enable && corepack prepare pnpm@${PNPM_VERSION} --activate
WORKDIR /src

# Manifests first so the dependency layer is cached.
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY frontend/package.json frontend/
COPY migration/package.json migration/
COPY openapi/package.json openapi/
COPY proto/gen/typescript/package.json proto/gen/typescript/
RUN --mount=type=cache,id=pnpm,target=/pnpm/store \
    pnpm install --frozen-lockfile --filter frontend...

# Generated proto TS sources (committed); dist/ is built here, then the frontend sources.
COPY proto/gen/typescript proto/gen/typescript
COPY frontend frontend
RUN pnpm --filter proto build && pnpm --filter frontend build

# ---- Backend build: embeds the frontend (core/internal/app/frontend.go) ----
FROM golang:${GO_VERSION} AS core
WORKDIR /src

COPY go.work go.work.sum ./
COPY core/go.mod core/
COPY proto/gen/go proto/gen/go
RUN --mount=type=cache,target=/go/pkg/mod \
    cd core && go mod download

COPY core core
# Replace the placeholder dist with the real build output before `go:embed all:dist`.
RUN rm -rf core/internal/app/dist
COPY --from=frontend /src/frontend/dist core/internal/app/dist

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd core && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/main ./cmd

# ---- Runtime ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=core /out/main /app/main

ENV LISTEN_ADDR=0.0.0.0:8080 \
    APP_ENV=production
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/main", "healthcheck"]

ENTRYPOINT ["/app/main"]
CMD ["serve"]
