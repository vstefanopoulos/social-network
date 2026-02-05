FROM node:20-slim AS base

# Install dependencies only when needed
FROM base AS deps
WORKDIR /app/social-sphere

# Install dependencies based on the preferred package manager
COPY social-sphere/package.json social-sphere/yarn.lock* social-sphere/package-lock.json* social-sphere/pnpm-lock.yaml* ./
RUN \
    if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
    elif [ -f package-lock.json ]; then npm ci; \
    elif [ -f pnpm-lock.yaml ]; then yarn global add pnpm && pnpm i --frozen-lockfile; \
    else echo "Lockfile not found." && exit 1; fi

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app/social-sphere

COPY social-sphere/ .
COPY --from=deps /app/social-sphere/node_modules ./node_modules

# Disable telemetry during build
#ENV NEXT_TELEMETRY_DISABLED=1

RUN npm run build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app/social-sphere

ENV NODE_ENV="production"
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs && \
    mkdir .next && \
    chown nextjs:nodejs .next

# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
# Order matters: standalone first, then static and public on top
COPY --from=builder --chown=nextjs:nodejs /app/social-sphere/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/social-sphere/.next/static ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/social-sphere/public ./public

USER nextjs

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]