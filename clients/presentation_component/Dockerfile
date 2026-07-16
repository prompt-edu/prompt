FROM node:24.14.0-alpine AS builder
WORKDIR /app
COPY package.json yarn.lock .yarnrc.yml ./
COPY presentation_component/package.json ./presentation_component/package.json
RUN corepack enable && yarn workspaces focus presentation_component
COPY presentation_component ./presentation_component
COPY tailwind-base.css ./tailwind-base.css
RUN yarn workspace presentation_component build

FROM nginx:1.29.6-alpine
COPY --from=builder /app/presentation_component/build /usr/share/nginx/html
EXPOSE 80
