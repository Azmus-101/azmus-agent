# ===== STAGE 1: Install ALL dependencies (prod + dev) =====
FROM node:lts-alpine AS deps
 
WORKDIR /app
 
COPY package*.json ./
RUN npm ci
 
 
# ===== STAGE 2: Build (compile TypeScript + generate Prisma) =====
# Used by CI — has compiled code AND dev dependencies (jest, ts-node, etc.)
FROM deps AS builder
 
COPY . .
 
RUN npm run prisma:generate
RUN NODE_OPTIONS="--max-old-space-size=3072" npm run build 
 
# ===== STAGE 3: Production runtime (lean image) =====
# Only compiled artifacts + prod dependencies — no dev tools, no source code
FROM node:lts-alpine AS production
 
WORKDIR /app
 
COPY package*.json ./
RUN npm ci --omit=dev
 
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/prisma ./prisma
COPY --from=builder /app/node_modules/.prisma ./node_modules/.prisma
COPY --from=builder /app/node_modules/@prisma ./node_modules/@prisma
COPY --from=builder /app/serviceAccounts.json ./serviceAccounts.json
 
EXPOSE 5000
 
CMD ["node", "dist/src/server.js"]
