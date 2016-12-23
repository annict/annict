# frozen_string_literal: true

source "https://rubygems.org"

ruby "2.3.3"

gem "rails", "~> 5.0.0"

gem "aasm"
gem "action_args"
gem "active_link_to"
gem "activerecord-session_store"
gem "acts_as_list"
gem "annotate"
gem "asset_sync"
gem "aws-sdk"
gem "bootstrap"
gem "bourbon"
gem "browser"
gem "browserify-rails"
gem "by_star"
gem "coffee-rails"
gem "delayed_job_active_record"
gem "devise"
gem "doorkeeper", ">= 4.2.0"
gem "draper", ">= 3.0.0.pre1"
gem "email_validator"
gem "enumerize"
gem "figaro"
gem "flutie"
gem "font-awesome-sass"
gem "gon"
gem "groupdate"
gem "hashdiff"
gem "httparty"
gem "imgix-rails"
gem "jb"
gem "jquery-rails"
gem "kaminari"
gem "keen"
gem "koala"
gem "meta-tags"
gem "mini_magick"
gem "nokogiri"
gem "omniauth-facebook"
# 1.4系だとFacebookのOAuth周りでおかしくなるので1.3系を使う
# https://github.com/intridea/omniauth-oauth2/issues/81
gem "omniauth-oauth2", "~> 1.3.1"
gem "omniauth-twitter"
gem "paperclip"
gem "pg"
gem "puma"
gem "puma_worker_killer"
gem "pundit"
gem "rack-cors", require: "rack/cors"
gem "rack-rewrite"
gem "rails_autolink"
gem "rails-html-sanitizer"
gem "rails-i18n"
gem "ransack"
gem "redis-rails"
gem "rmagick"
# To use font-awesome-sass
# https://github.com/sass/sassc-rails/issues/6
gem "sass-rails", require: false
gem "sassc-rails"
gem "sitemap_generator"
gem "slack-notifier"
gem "slim"
gem "traceroute"
gem "twitter"
gem "uglifier"
gem "validate_url"
gem "virtus"

group :development, :test do
  gem "awesome_print"
  gem "dmmyix"
  gem "hirb-unicode-steakknife"
  gem "hirb"
  gem "pry-alias"
  gem "pry-byebug"
  gem "pry-coolline"
  gem "pry-rails"
  gem "rails-flog", require: "flog"
  gem "rspec-rails"
  gem "rspec-mocks"
end

group :development do
  gem "active_record_query_trace"
  gem "better_errors"
  gem "binding_of_caller" # better_errorsで使用
  gem "bullet"
  gem "derailed_benchmarks"
  gem "letter_opener_web"
  gem "listen" # Rails 5から `rails s` するときに必要になった
  gem "rubocop"
  gem "ruby_identicon"
  gem "scss_lint", require: false
  gem "spring"
  gem "spring-commands-rspec", require: false
  gem "stackprof"
  gem "thin"
end

group :test do
  gem "capybara"
  gem "coveralls", require: false
  gem "database_rewinder"
  gem "factory_girl_rails"
  gem "poltergeist"
end

group :production do
  gem "bugsnag"
  gem "rails_12factor"
  gem "scout_apm"
end

source "https://rails-assets.org" do
  gem "rails-assets-tether"
end
