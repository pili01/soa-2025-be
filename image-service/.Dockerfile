FROM node:20-alpine

WORKDIR /usr/src/app

COPY package*.json ./

RUN npm install --production

COPY . .

RUN mkdir -p uploads/ProfilePictures

EXPOSE 3000

CMD ["node", "src/app.js"]