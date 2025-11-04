# ---------- Builder Stage ----------
FROM node:22.18.0 AS builder

WORKDIR /app

RUN mkdir -p ./frontend/messages
RUN apt update \
 && apt install -y --no-install-recommends protobuf-compiler

# Copy root-level package/config files to install dependencies
COPY package*.json ./
COPY nuxt.config.ts ./
COPY tailwind.config.js ./
COPY tsconfig.json ./
COPY components.json ./

# Install dependencies
RUN npm ci

COPY protobufs/ ./protobufs/
RUN protoc --plugin=protoc-gen-ts_proto=./node_modules/.bin/protoc-gen-ts_proto \
  --ts_proto_opt=esModuleInterop=true,forceLong=string \
  --ts_proto_out=frontend/messages \
  protobufs/*.proto

# Copy the frontend source code
COPY frontend/ ./frontend/

# Set working directory to srcDir
WORKDIR /app/frontend

# Build Nuxt app
RUN npm run build

# ---------- Production Stage ----------
FROM node:22.18.0-alpine AS production

WORKDIR /app

# Copy the built Nuxt output from builder
COPY --from=builder /app/.output ./.output

# Expose Nuxt port
EXPOSE 3000

# Start Nuxt server
CMD ["node", ".output/server/index.mjs"]
