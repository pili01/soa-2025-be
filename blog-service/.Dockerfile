FROM node:18-alpine AS builder
WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .
RUN npx prisma generate

# ------- RUNTIME --------
FROM node:18-alpine
WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production


COPY --from=builder /app/node_modules/.prisma ./node_modules/.prisma

COPY --from=builder /app/node_modules/prisma ./node_modules/prisma
COPY --from=builder /app/node_modules/.bin/prisma ./node_modules/.bin/prisma

COPY --from=builder /app/prisma ./prisma

COPY ./src ./src

COPY --from=builder /app/docker-entrypoint.sh /usr/local/bin/entrypoint.sh
RUN sed -i 's/\r$//' /usr/local/bin/entrypoint.sh && chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/bin/sh", "/usr/local/bin/entrypoint.sh"]
EXPOSE 3000
CMD [ "node", "src/app.js" ]
