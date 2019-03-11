FROM node:10.15.1-alpine as node
FROM ruby:2.6.1-alpine

LABEL maintainer="https://annict.jp/@shimbaco" \
      description="A platform for anime addicts."

RUN apk update && \
    apk add -t \
        build-dependencies \
        build-base \
        git \
        postgresql \
        postgresql-dev \
        imagemagick \
        yarn

ENV PATH=./node_modules/.bin/:$PATH \
    RAILS_ENV=development

# Set to install cld gem
# https://github.com/jtoy/cld/issues/10
ENV CFLAGS=-Wno-narrowing \
    CXXFLAGS=-Wno-narrowing
# Set to run `ls` in Pry
# https://github.com/pry/pry/issues/1494#issuecomment-162336567
ENV PAGER=busybox\ less

WORKDIR /annict/

COPY --from=node /usr/local/bin/node /usr/local/bin/

COPY Gemfile* package.json yarn.lock ./
RUN gem install bundler && \
    bundle install -j$(getconf _NPROCESSORS_ONLN) && \
    yarn install && \
    yarn cache clean

EXPOSE 3000
