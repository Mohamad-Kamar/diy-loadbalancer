FROM node:20-alpine
WORKDIR /app
COPY index.js .
RUN npm init -y && npm install express
EXPOSE 8082
CMD ["node", "index.js"]
