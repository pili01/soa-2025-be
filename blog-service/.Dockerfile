# ---------- BUILDER ----------
FROM node:18-bullseye AS builder
WORKDIR /app

# Instaliraj sve dependencies
COPY package*.json ./
RUN npm install

# Kopiraj sve fajlove
COPY . .

# Generiši Prisma client u builderu
RUN npx prisma generate

# ---------- RUNTIME ----------
FROM node:18-bullseye-slim
WORKDIR /app

# Instaliraj samo production dependencies
COPY package*.json ./
RUN npm ci --only=production

# Kopiraj node_modules i Prisma binarije iz buildera
COPY --from=builder /app/node_modules/.prisma ./node_modules/.prisma
COPY --from=builder /app/node_modules/prisma ./node_modules/prisma
COPY --from=builder /app/node_modules/.bin/prisma ./node_modules/.bin/prisma

# Kopiraj Prisma schema
COPY --from=builder /app/prisma ./prisma

# Kopiraj source
COPY ./src ./src

# Kopiraj entrypoint
COPY --from=builder /app/docker-entrypoint.sh /usr/local/bin/entrypoint.sh
RUN sed -i 's/\r$//' /usr/local/bin/entrypoint.sh && chmod +x /usr/local/bin/entrypoint.sh

# Postavi entrypoint i expose port
ENTRYPOINT ["/bin/sh", "/usr/local/bin/entrypoint.sh"]
EXPOSE 3000
CMD ["node", "src/app.js"]
