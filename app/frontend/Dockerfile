FROM node:10.15.1-alpine

LABEL maintainer="https://annict.jp/@shimbaco" \
      description="A platform for anime addicts."

RUN apk update && \
    apk add -t \
        git

ENV PATH=./node_modules/.bin/:$PATH \
    NODE_ENV=development

WORKDIR /annict/

COPY package.json yarn.lock webpack.config.js ./
RUN yarn install && \
    yarn cache clean

EXPOSE 8080
