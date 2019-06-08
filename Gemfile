# frozen_string_literal: true

source "https://rubygems.org"
git_source(:github) { |repo| "https://github.com/#{repo}.git" }

ruby "2.6.3"

gem "rails", github: "rails/rails"

gem "bootsnap", github: "Shopify/bootsnap", require: false
gem "pg", github: "ged/ruby-pg"
gem "puma", github: "puma/puma"

group :development, :test do
  gem "dotenv-rails", github: "bkeepers/dotenv"
  gem "pry", github: "pry/pry"
  gem "pry-rails", github: "rweng/pry-rails"
end

group :development do
  gem "listen", github: "guard/listen"
  gem "spring", github: "rails/spring"
  gem "spring-watcher-listen", github: "jonleighton/spring-watcher-listen"
end
