# syntax=docker/dockerfile:1

ARG NODE_VERSION=20
ARG APP_PORT=3000
ARG NODE_UID=1001
ARG NODE_GID=1001

# ── Stage 1: deps ──────────────────────────────────────────────
FROM node:${NODE_VERSION}-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

# ── Stage 2: builder ───────────────────────────────────────────
FROM node:${NODE_VERSION}-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ARG NEXT_PUBLIC_API_URL
ARG NEXT_PUBLIC_FRONTEND_DOMAIN
ARG NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT

ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_FRONTEND_DOMAIN=$NEXT_PUBLIC_FRONTEND_DOMAIN
ENV NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT=$NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# ── Stage 3: runner ────────────────────────────────────────────
FROM node:${NODE_VERSION}-alpine AS runner
WORKDIR /app

ARG APP_PORT
ARG NODE_UID
ARG NODE_GID

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid ${NODE_GID} nodejs \
    && adduser --system --uid ${NODE_UID} nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs
EXPOSE ${APP_PORT}
ENV PORT=${APP_PORT}
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
