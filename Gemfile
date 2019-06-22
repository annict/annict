# frozen_string_literal: true

source "https://rubygems.org"
git_source(:github) { |repo| "https://github.com/#{repo}.git" }

ruby "2.6.3"

gem "rails", github: "rails/rails"

gem "bootsnap", github: "Shopify/bootsnap", require: false
gem "devise", github: "shimbaco/devise", branch: "rails-6.1"
gem "email_validator", github: "balexand/email_validator"
gem "pg", github: "ged/ruby-pg"
gem "puma", github: "puma/puma"
gem "sorbet-runtime"

group :development, :test do
  gem "dotenv-rails", github: "bkeepers/dotenv"
  gem "pry", github: "pry/pry"
  gem "pry-rails", github: "rweng/pry-rails"
  gem "rspec-core", github: "rspec/rspec-core"
  gem "rspec-expectations", github: "rspec/rspec-expectations"
  gem "rspec-mocks", github: "rspec/rspec-mocks"
  gem "rspec-rails", github: "rspec/rspec-rails"
  gem "rspec-support", github: "rspec/rspec-support"
  gem "rspec_junit_formatter", github: "sj26/rspec_junit_formatter" # Using on CircleCI
end

group :development do
  gem "listen", github: "guard/listen"
  gem "sorbet"
  gem "spring", github: "rails/spring"
  gem "spring-watcher-listen", github: "jonleighton/spring-watcher-listen"
end

group :test do
  gem "capybara", github: "teamcapybara/capybara"
  gem "factory_bot_rails", github: "thoughtbot/factory_bot_rails"
  gem "selenium-webdriver"
  gem "simplecov", github: "colszowka/simplecov", require: false
  gem "webdrivers", github: "titusfortner/webdrivers"
end
