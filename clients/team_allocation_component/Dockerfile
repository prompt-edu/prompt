ARG IMAGE_TAG=latest
FROM ghcr.io/prompt-edu/prompt/prompt-clients-base:${IMAGE_TAG} AS core-base


WORKDIR /app/team_allocation_component
COPY . ./

ARG NODE_ENV=production
RUN yarn webpack --mode=${NODE_ENV} --env NODE_ENV=${NODE_ENV}

# Build the final image
FROM nginx:stable-alpine

COPY --from=core-base /app/team_allocation_component/build /usr/share/nginx/html
COPY --from=core-base /app/nginx/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
