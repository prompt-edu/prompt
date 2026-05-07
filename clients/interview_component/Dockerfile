ARG IMAGE_TAG=latest
FROM ghcr.io/prompt-edu/prompt/prompt-clients-base:${IMAGE_TAG} AS core-base

ARG CORE_HOST
ENV REACT_APP_CORE_HOST=$CORE_HOST

WORKDIR /app/interview_component
COPY . ./

ARG NODE_ENV=production
RUN yarn webpack --mode=${NODE_ENV} --env NODE_ENV=${NODE_ENV}

# Build the final image
FROM nginx:stable-alpine

COPY --from=core-base /app/interview_component/build /usr/share/nginx/html
COPY --from=core-base /app/nginx/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
