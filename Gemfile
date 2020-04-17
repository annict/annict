# frozen_string_literal: true

source "https://rubygems.org"
git_source(:github) { |repo| "https://github.com/#{repo}.git" }

ruby "2.7.1"

gem "rails", "~> 6.0.0"

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
gem "delayed_job_active_record"
gem "devise"
gem "discord-notifier"
gem "doorkeeper"
gem "down"
gem "dry-struct"
gem "email_validator"
gem "enumerize"
gem "flutie"
gem "github-markup"
gem "gon"
gem "graphql", ">= 1.10.0.pre3" # https://github.com/rmosolgo/graphql-ruby/pull/2640
gem "graphql-batch"
gem "gretel"
gem "groupdate"
gem "hashdiff"
gem "htmlrb"
gem "http_accept_language"
gem "httparty"
gem "image_processing"
gem "imgix-rails"
gem "jb"
gem "kaminari"
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
gem "pg"
gem "puma"
gem "puma_worker_killer"
gem "pundit"
gem "rack-attack"
gem "rack-cors", require: "rack/cors"
gem "rack-rewrite"
gem "rails-html-sanitizer"
gem "rails-i18n"
gem "rails_autolink"
gem "ransack"
gem "sentry-raven"
gem "shrine"
gem "sitemap_generator"
gem "slim"
gem "syoboi_calendar"
gem "twitter"
gem "validate_url"
gem "view_component"
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
  gem "factory_bot_rails"
  gem "selenium-webdriver"
  # Use < 0.18 until the following issue will be resolved.
  # https://github.com/codeclimate/test-reporter/issues/418
  gem "simplecov", "< 0.18", require: false
  gem "timecop"
  gem "webdrivers", require: !ENV["CI"] # Added to run spec with Chrome on local machine
end

group :production do
  gem "lograge"
end
