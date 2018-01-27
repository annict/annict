# frozen_string_literal: true

source "https://rubygems.org"

ruby "2.5.0"

git_source(:github) do |repo_name|
  repo_name = "#{repo_name}/#{repo_name}" unless repo_name.include?("/")
  "https://github.com/#{repo_name}.git"
end

gem "rails", "5.1.4"

gem "aasm"
gem "action_args"
gem "active_link_to"
gem "activerecord-session_store"
gem "acts_as_list"
gem "amazon-ecs"
gem "annotate"
# Use aws-sdk 2.x for Paperclip
# https://github.com/thoughtbot/paperclip/issues/2484
gem "aws-sdk", "< 3.0"
gem "bootsnap", require: false
gem "browser"
gem "by_star"
gem "cld"
gem "commonmarker"
gem "counter_culture"
gem "dalli"
gem "delayed_job"
gem "delayed_job_active_record"
gem "devise"
gem "discord-notifier"
gem "doorkeeper"
gem "draper"
gem "email_validator"
gem "enumerize"
gem "figaro"
gem "flutie"
gem "github-markup"
gem "gon"
gem "graphql"
gem "graphql-batch"
gem "gretel"
gem "groupdate"
gem "hashdiff"
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
gem "rmagick"
gem "sentry-raven"
gem "sitemap_generator"
gem "slack-notifier"
gem "slim"
gem "traceroute"
gem "twitter"
gem "validate_url"
gem "virtus"
gem "webpacker"
gem "wilson_score"

group :development, :test do
  gem "awesome_print"
  gem "dmmyix"
  gem "pry"
  gem "pry-alias"
  gem "pry-byebug"
  gem "pry-coolline"
  gem "pry-rails"
  gem "rspec-mocks"
  gem "rspec-rails"
end

group :development do
  gem "active_record_query_trace"
  gem "better_errors"
  gem "binding_of_caller" # Using better_errors
  gem "bullet"
  gem "derailed_benchmarks"
  gem "fast_stack" # Using rack-mini-profiler
  gem "flamegraph" # Using rack-mini-profiler
  gem "graphiql-rails"
  gem "i18n-tasks"
  gem "letter_opener_web"
  gem "listen" # Rails 5から `rails s` するときに必要になった
  gem "memory_profiler" # Using rack-mini-profiler
  gem "rack-mini-profiler", require: false
  gem "rubocop"
  gem "ruby_identicon"
  gem "scss_lint", require: false
  gem "spring-commands-rspec", require: false
  gem "spring"
  gem "squasher"
  gem "stackprof" # Using rack-mini-profiler
  gem "thin"
end

group :test do
  gem "capybara"
  gem "coveralls", require: false
  gem "database_rewinder"
  gem "factory_bot_rails"
  gem "poltergeist"
  gem "timecop"
end

group :production do
  gem "bugsnag"
  gem "heroku-deflater"
  gem "lograge"
end
