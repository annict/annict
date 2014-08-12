source 'https://rubygems.org'

ruby '2.1.1'

gem 'rails', '4.1.4'

gem 'action_args'
gem 'activerecord-session_store'
gem 'asset_sync'
gem 'bootstrap-sass'
gem 'bourbon', '3.2.3' # sass-railsがsass 3.3系をサポートするまで3系を使用する
gem 'browser'
gem 'coffee-rails'
gem 'devise'
gem 'dragonfly-s3_data_store'
gem 'email_validator'
gem 'enumerize'
gem 'exception_notification'
gem 'figaro'
gem 'flutie'
gem 'font-awesome-sass'
gem 'foreigner'
gem 'gon'
gem 'jbuilder'
gem 'jquery-rails'
gem 'kaminari'
gem 'koala'
gem 'mysql2'
gem 'newrelic_rpm'
gem 'ngannotate-rails'
gem 'nokogiri'
gem 'omniauth-facebook'
gem 'omniauth-twitter'
gem 'paper_trail'
gem 'rails-i18n'
gem 'ransack'
gem 'sass-rails'
gem 'sidekiq'
gem 'sidekiq-middleware' # Recommendableで使用
gem 'slim'
gem 'recommendable' # gem 'sidekiq' より下に置く必要があるらしい
gem 'twitter'
gem 'uglifier'
gem 'unicorn'
gem 'whenever', require: false

group :development, :test do
  gem 'awesome_print'
  gem 'hirb-unicode'
  gem 'hirb'
  gem 'pry-byebug'
  gem 'pry-rails'
  gem 'rails-flog', require: 'flog'
  gem 'rspec-rails'
end

group :development do
  gem 'aws-sdk'
  gem 'better_errors'
  gem 'binding_of_caller' # using better_errors
  gem 'bullet'
  gem 'capistrano'
  gem 'capistrano-rails'
  gem 'capistrano-rvm'
  gem 'capistrano-sidekiq'
  gem 'capistrano-unicorn-nginx'
  gem 'letter_opener'
  gem 'meta_request'
  gem 'quiet_assets'
  gem 'spring-commands-rspec', require: false
  gem 'spring'
  gem 'thin'
end

group :test do
  gem 'capybara'
  gem 'coveralls', require: false
  gem 'database_rewinder'
  gem 'factory_girl_rails'
  gem 'nyan-cat-formatter'
  gem 'poltergeist'
end

group :production do
  gem 'skylight'
end