# Build stage: full workspace (web needs packages/contracts + the pnpm workspace).
FROM node:22-slim AS build
RUN npm install -g pnpm@12.4.1
WORKDIR /app

# Vite inlines these at build time; Coolify injects app env vars into builds.
ARG VITE_API_URL
ARG VITE_CLERK_PUBLISHABLE_KEY
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_CLERK_PUBLISHABLE_KEY=$VITE_CLERK_PUBLISHABLE_KEY

COPY pnpm-workspace.yaml pnpm-lock.yaml ./
COPY packages/contracts ./packages/contracts
COPY apps/web ./apps/web
RUN pnpm install --frozen-lockfile

WORKDIR /app/apps/web
RUN pnpm compile:messages && pnpm build

# Runtime stage: workspace + built output; serve via the app's own start script.
FROM node:22-slim AS runtime
RUN npm install -g pnpm@12.4.1
WORKDIR /app
COPY --from=build /app /app

ENV NODE_ENV=production
EXPOSE 3000
CMD ["pnpm", "--filter", "web", "start"]
