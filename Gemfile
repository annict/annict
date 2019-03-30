# frozen_string_literal: true

source "https://rubygems.org"
git_source(:github) { |repo| "https://github.com/#{repo}.git" }

ruby "2.6.2"

gem "rails"

gem "aasm"
gem "active_decorator"
gem "active_link_to"
gem "acts_as_list"
gem "amazon-ecs"
gem "aws-sdk-s3" # Uses in Paperclip
gem "bootsnap", require: false
gem "browser", require: "browser/browser"
gem "by_star"
gem "cld"
gem "commonmarker"
gem "counter_culture"
gem "devise"
gem "discord-notifier"
gem "doorkeeper"
gem "email_validator"
gem "enumerize"
gem "flutie"
gem "github-markup"
gem "gon"
gem "graphql"
gem "graphql-batch"
gem "gretel"
gem "groupdate"
gem "hashdiff"
gem "hiredis"
gem "http_accept_language"
gem "httparty"
gem "imgix-rails"
gem "impressionist"
gem "jb"
gem "kaminari"
gem "keen"
gem "koala"
gem "meta-tags"
gem "mini_magick"
gem "miro"
gem "mjml-rails"
gem "moji"
gem "nokogiri"
gem "omniauth-facebook"
gem "omniauth-gumroad"
gem "omniauth-twitter"
gem "paperclip"
gem "pg"
gem "puma"
gem "puma_worker_killer"
gem "pundit"
gem "rack-cors", require: "rack/cors"
gem "rack-rewrite"
gem "rails-html-sanitizer"
gem "rails-i18n"
gem "rails_autolink"
gem "ransack"
gem "redis"
gem "redis-session-store"
gem "sentry-raven"
gem "sidekiq"
gem "sitemap_generator"
gem "slim"
gem "syoboi_calendar"
gem "twitter"
gem "validate_url"
gem "virtus"
gem "wilson_score"

group :development, :test do
  gem "awesome_print"
  gem "dotenv-rails"
  gem "pry-rails"
  gem "rspec-mocks"
  gem "rspec-rails"
  gem "rspec_junit_formatter" # Using on CircleCI
end

group :development do
  gem "active_record_query_trace"
  gem "annotate"
  gem "better_errors"
  gem "binding_of_caller" # Using better_errors
  gem "bullet"
  gem "derailed_benchmarks"
  gem "graphiql-rails"
  gem "graphql-docs"
  gem "i18n-tasks"
  gem "listen" # Using with `rails s` since Rails 5
  gem "memory_profiler"
  gem "meta_request"
  gem "rubocop"
  gem "ruby_identicon"
  gem "scss_lint", require: false
  gem "solargraph"
  gem "spring-commands-rspec", require: false
  gem "spring"
  gem "squasher"
  gem "stackprof"
  gem "traceroute"
end

group :test do
  gem "capybara"
  gem "chromedriver-helper"
  gem "database_rewinder"
  gem "factory_bot_rails"
  gem "selenium-webdriver"
  gem "simplecov", require: false
  gem "timecop"
end

group :production do
  gem "lograge"
end
