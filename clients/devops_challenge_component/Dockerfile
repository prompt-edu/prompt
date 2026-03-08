ARG IMAGE_TAG=latest
FROM ghcr.io/prompt-edu/prompt/prompt-clients-base:${IMAGE_TAG} AS core-base


WORKDIR /app/devops_challenge_component
COPY . ./
RUN yarn install

RUN yarn build

# Build the final image
FROM nginx:stable-alpine

COPY --from=core-base /app/devops_challenge_component/build /usr/share/nginx/html
COPY --from=core-base /app/nginx/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]