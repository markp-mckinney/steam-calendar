FROM node:22-alpine
WORKDIR /app
COPY package*.json ./
ENV NODE_ENV=production
RUN npm install
COPY src src        
CMD npm start