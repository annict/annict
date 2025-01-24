# frozen_string_literal: true

source "https://rubygems.org"
git_source(:github) { |repo| "https://github.com/#{repo}.git" }

ruby "3.3.5"

gem "rails", "~> 7.0.8"

gem "active_decorator"
gem "active_link_to"
gem "activerecord-session_store"
gem "acts_as_list"
gem "addressable"
gem "aws-sdk-s3" # Shrineで使用
gem "browser", require: "browser/browser"
gem "by_star"
gem "cld"
# github-markupが1系に対応するまで0系を使う
# ref: https://github.com/github/markup/issues/1758
gem "commonmarker", "< 3.0" # Using github-markup
gem "counter_culture"
gem "cssbundling-rails"
gem "delayed_job_active_record"
gem "devise"
gem "discord-notifier"
gem "doorkeeper"
gem "dotenv-rails"
gem "down"
gem "dry-struct"
gem "email_validator"
gem "enumerize"
gem "github-markup"
gem "graphql", "~> 2.0.31"
gem "graphql-batch"
gem "graphql-fragment_cache"
gem "groupdate"
gem "hashdiff"
gem "hiredis"
gem "htmlrb", github: "shimbaco/htmlrb", branch: "main"
gem "http_accept_language"
gem "httparty"
gem "icalendar"
gem "image_processing"
gem "imgproxy"
gem "jb"
gem "jsbundling-rails"
gem "kaminari"
gem "koala"
gem "lograge"
gem "memory_profiler" # Used by rack-mini-profiler
gem "meta-tags"
gem "mini_magick"
gem "mini_mime"
gem "mjml-rails"
gem "moji"
gem "nokogiri"
gem "omniauth-facebook"
gem "omniauth-gumroad"
gem "omniauth-rails_csrf_protection"
gem "pg"
gem "prelude-batch-loader", require: "prelude"
gem "propshaft"
gem "puma"
gem "puma_worker_killer"
gem "pundit"
gem "rack-attack"
gem "rack-cors", require: "rack/cors"
gem "rack-mini-profiler"
gem "rack-rewrite"
gem "rack-timeout", require: "rack/timeout/base"
gem "rails_autolink"
gem "rails-html-sanitizer"
gem "rails-i18n"
gem "ransack"
gem "redis"
gem "sentry-ruby"
gem "sentry-rails"
gem "shrine"
gem "slim"
gem "sorbet-runtime"
gem "syoboi_calendar"
gem "view_component"
gem "virtus"
gem "wilson_score"

group :development, :test do
  gem "awesome_print"
  gem "erb_lint", require: false
  gem "factory_bot_rails"
  gem "pry-rails"
  gem "rspec-mocks"
  gem "rspec-rails"
  gem "rubocop-factory_bot", require: false
  gem "rubocop-rspec", require: false
  gem "standard"
  gem "standard-rails"
  gem "standard-sorbet"
end

group :development do
  gem "active_record_query_trace"
  gem "better_errors"
  gem "binding_of_caller" # Using better_errors
  gem "bullet"
  # Temporary comment out until graphql-docs will replace sass with sassc.
  # The sass gem causes LoadError.
  # https://github.com/gjtorikian/graphql-docs/issues/86
  # gem "graphql-docs"
  gem "i18n-tasks"
  gem "listen" # Using with `rails s` since Rails 5
  gem "solargraph"
  gem "sorbet"
  gem "squasher"
  gem "tapioca", require: false
  gem "traceroute"
end

group :test do
  # Use < 0.18 until the following issue will be resolved.
  # https://github.com/codeclimate/test-reporter/issues/418
  gem "simplecov", "< 0.22", require: false
  gem "timecop"
end

group :production do
  gem "resend"
end
