# frozen_string_literal: true

source "https://rubygems.org"
git_source(:github) { |repo| "https://github.com/#{repo}.git" }

ruby "3.1.2"

gem "rails", "~> 7.0.0"

gem "active_decorator"
gem "active_link_to"
gem "activerecord-session_store"
gem "acts_as_list"
gem "addressable"
gem "aws-sdk-s3" # Using in Shrine
gem "browser", require: "browser/browser"
gem "by_star"
gem "cld"
gem "commonmarker" # Using github-markup
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
gem "graphql", ">= 1.10.0.pre3" # https://github.com/rmosolgo/graphql-ruby/pull/2640
gem "graphql-batch"
gem "graphql-fragment_cache"
gem "groupdate"
gem "hashdiff"
gem "hiredis"
gem "htmlrb", github: "kiraka/htmlrb", branch: "main"
gem "http_accept_language"
gem "httparty"
gem "image_processing"
gem "imgproxy"
gem "jb"
gem "jsbundling-rails"
gem "kaminari"
gem "koala"
gem "memory_profiler" # Used by rack-mini-profiler
gem "meta-tags"
gem "mini_magick"
gem "mjml-rails"
gem "moji"
gem "nokogiri"
gem "omniauth-facebook"
gem "omniauth-gumroad"
gem "omniauth-rails_csrf_protection"
gem "omniauth-twitter"
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
gem "rails_autolink"
gem "rails-html-sanitizer"
gem "rails-i18n"
gem "ransack"
gem "redis"
gem "sentry-ruby"
gem "sentry-rails"
gem "shrine"
gem "slim"
gem "syoboi_calendar"
gem "twitter"
gem "view_component"
gem "virtus"
gem "wilson_score"

group :development, :test do
  gem "awesome_print"
  gem "factory_bot_rails"
  gem "pry-rails"
  gem "rspec-mocks"
  gem "rspec-rails"
  gem "standard"
end

group :development do
  gem "active_record_query_trace"
  gem "annotate"
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
  gem "squasher"
  gem "traceroute"
end

group :test do
  # Use < 0.18 until the following issue will be resolved.
  # https://github.com/codeclimate/test-reporter/issues/418
  gem "simplecov", "< 0.22", require: false
  gem "timecop"
end
