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
        nodejs-npm \
        imagemagick

ENV RAILS_ENV=development

# Set to install cld gem
# https://github.com/jtoy/cld/issues/10
ENV CFLAGS=-Wno-narrowing \
    CXXFLAGS=-Wno-narrowing
# Set to run `ls` in Pry
# https://github.com/pry/pry/issues/1494#issuecomment-162336567
ENV PAGER=busybox\ less

WORKDIR /annict/

COPY Gemfile* ./
RUN bundle install -j$(getconf _NPROCESSORS_ONLN) && \
    npm i -g mjml@4.1.2

EXPOSE 3000
