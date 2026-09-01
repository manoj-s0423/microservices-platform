FROM node:20-alpine

WORKDIR /app

COPY package.json package-lock.json ./

RUN npm ci 

COPY src ./src

USER node

ENV NODE_ENV=production

EXPOSE 8084

# Start the service. The `node` user is already set as the default user, so no need to `USER node` again here. The CMD is the last instruction in the Dockerfile, so it doesn't need to be `exec` form (JSON array) — but it is a good habit to use that form anyway, so that signals are forwarded properly to the Node.js process and it can exit gracefully on SIGTERM.
CMD ["node" , "dist/index.js"]