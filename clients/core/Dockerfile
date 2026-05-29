# base will be the the image build by the base dockerfile.
ARG IMAGE_TAG=latest
FROM ghcr.io/prompt-edu/prompt/prompt-clients-base:${IMAGE_TAG} AS core-base

WORKDIR /app/core

COPY . ./

ARG NODE_ENV=production
RUN yarn webpack --mode=${NODE_ENV} --env NODE_ENV=${NODE_ENV}

FROM nginx:stable-alpine

COPY --from=core-base /app/core/build /usr/share/nginx/html
COPY --from=core-base /app/nginx/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=core-base /app/core/build/env.template.js /usr/share/nginx/html/env.template.js

# Copy the entrypoint script to the image.
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 80

# Use the custom entrypoint to generate env.js and start Nginx.
ENTRYPOINT ["/docker-entrypoint.sh"]
