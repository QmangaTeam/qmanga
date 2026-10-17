FROM oven/bun@sha256:629e17411f1f129dbec3af78d5af9c9f2a937435c80349437206c6b0b7422373 AS build
WORKDIR /app
COPY apps/web/package.json apps/web/bun.lock ./
RUN bun install --frozen-lockfile
COPY apps/web/ .
RUN bun run check:architecture && bun run check && bun run build
FROM oven/bun@sha256:629e17411f1f129dbec3af78d5af9c9f2a937435c80349437206c6b0b7422373
WORKDIR /app
COPY --from=build --chown=bun:bun /app/build ./build
COPY --from=build --chown=bun:bun /app/package.json ./package.json
ENV HOST=0.0.0.0 PORT=3000
USER bun
EXPOSE 3000
CMD ["bun", "build/index.js"]

