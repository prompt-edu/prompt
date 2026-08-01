# Self-contained build (no prompt-clients-base image needed outside the monorepo)
FROM node:22-alpine AS build

WORKDIR /app
# .yarnrc.yml is optional (Corepack reads the Yarn version from package.json);
# the [l] glob copies it when present but does not fail the build if absent.
COPY package.json yarn.lock .yarnrc.ym[l] ./
RUN corepack enable && yarn install --immutable

COPY . ./
RUN yarn build

FROM nginx:stable-alpine

COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
