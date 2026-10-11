# syntax=docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e

# Base images are pinned by digest for reproducible, tamper-evident builds; Dependabot bumps them.
ARG NODE_VERSION=24
ARG APP_PORT=3000
ARG NODE_UID=1001
ARG NODE_GID=1001

# ── Stage 1: deps ──────────────────────────────────────────────
FROM node:${NODE_VERSION}-alpine@sha256:ebfe2f90462722a7a4de65e91990e97fe0d401c70e0e762c5b53302f905ec1c1 AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

# ── Stage 2: builder ───────────────────────────────────────────
FROM node:${NODE_VERSION}-alpine@sha256:ebfe2f90462722a7a4de65e91990e97fe0d401c70e0e762c5b53302f905ec1c1 AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ARG NEXT_PUBLIC_API_URL
ARG NEXT_PUBLIC_FRONTEND_DOMAIN
ARG NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT
ARG NEXT_PUBLIC_FEEDBACK_URL

ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_FRONTEND_DOMAIN=$NEXT_PUBLIC_FRONTEND_DOMAIN
ENV NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT=$NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT
ENV NEXT_PUBLIC_FEEDBACK_URL=$NEXT_PUBLIC_FEEDBACK_URL
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# ── Stage 3: runner ────────────────────────────────────────────
FROM node:${NODE_VERSION}-alpine@sha256:ebfe2f90462722a7a4de65e91990e97fe0d401c70e0e762c5b53302f905ec1c1 AS runner
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
