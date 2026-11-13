# build: compila TypeScript y permite correr los tests.
FROM node:26.13.0-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY tsconfig.json ./
COPY src ./src
COPY test ./test
RUN npm run build
CMD ["npm", "test"]

# runtime: solo dependencias de producción y JS compilado, sin root.
FROM node:26.13.0-alpine AS runtime
ENV NODE_ENV=production
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev && npm cache clean --force
COPY --from=build /app/dist ./dist
USER node
EXPOSE 4000
CMD ["node", "dist/server.js"]
