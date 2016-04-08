# frozen_string_literal: true

source "https://rubygems.org"

ruby "2.3.0"

gem "rails", "~> 4.2.5.2"

gem "aasm"
gem "action_args"
gem "active_link_to"
gem "activerecord-session_store"
gem "acts_as_list"
gem "angular_rails_csrf"
gem "annotate"
gem "asset_sync"
gem "aws-sdk-v1" # Paperclipで使用
gem "bootstrap-sass"
gem "bourbon"
# 2系対応が面倒くさそうなので一旦1系を使い続ける
gem "browser", "~> 1.1.0"
gem "browserify-rails"
gem "by_star"
gem "coffee-rails"
gem "delayed_job_active_record"
gem "devise", ">= 3.5.4"
gem "draper"
gem "elasticsearch-model"
gem "elasticsearch-rails"
gem "email_validator"
gem "enumerize"
gem "figaro"
gem "flutie"
gem "font-awesome-sass"
gem "gon"
gem "hashdiff"
gem "httparty"
gem "imgix-rails"
gem "jbuilder"
gem "jquery-rails"
gem "kaminari"
gem "keen"
gem "koala"
gem "meta-tags"
gem "mini_magick"
gem "net-ssh" # fog を使用している asset_sync で使用
gem "ngannotate-rails"
gem "nokogiri", ">= 1.6.7.2"
gem "omniauth-facebook"
# 1.4系だとFacebookのOAuth周りでおかしくなるので1.3系を使う
# https://github.com/intridea/omniauth-oauth2/issues/81
gem "omniauth-oauth2", "~> 1.3.1"
gem "omniauth-twitter"
gem "paper_trail"
gem "paperclip"
gem "pg"
gem "puma"
gem "pundit"
gem "rack-rewrite"
gem "rails_autolink"
gem "rails-i18n"
gem "redis-rails"
gem "rmagick"
gem "sass-rails"
gem "sinatra", require: nil
gem "sitemap_generator"
gem "slim"
gem "traceroute"
gem "twitter"
gem "uglifier"
gem "validate_url"
gem "virtus"

group :development, :test do
  gem "awesome_print"
  gem "hirb-unicode-steakknife"
  gem "hirb"
  gem "pry-byebug"
  gem "pry-coolline"
  gem "pry-rails"
  gem "rails-flog", require: "flog"
  gem "rspec-rails"
  gem "rspec-mocks"
end

group :development do
  gem "aws-sdk"
  gem "better_errors"
  gem "binding_of_caller" # better_errorsで使用
  gem "bullet"
  gem "letter_opener_web"
  gem "quiet_assets"
  gem "rubocop"
  gem "ruby_identicon"
  gem "spring"
  gem "spring-commands-rspec", require: false
  gem "thin"
end

group :production do
  gem "bugsnag"
  gem "rails_12factor"
  gem "skylight"
end

group :test do
  gem "capybara"
  gem "codeclimate-test-reporter", require: nil
  gem "database_rewinder"
  gem "factory_girl_rails"
  gem "poltergeist"
end

gem "rails-html-sanitizer", "~> 1.0.3"
