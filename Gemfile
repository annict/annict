source 'https://rubygems.org'
source 'https://rails-assets.org'

ruby '2.1.5'

gem 'rails', '4.1.8'

gem 'action_args'
gem 'activerecord-session_store'
gem 'annotate'
gem 'asset_sync'
gem 'bootstrap-sass'
gem 'bourbon', '3.2.3' # sass-railsがsass 3.3系をサポートするまで3系を使用する
gem 'browser'
gem 'by_star'
gem 'coffee-rails'
gem 'devise'
gem 'dragonfly', '1.0.7'
gem 'dragonfly-s3_data_store'
gem 'email_validator'
gem 'enumerize'
gem 'figaro'
gem 'flutie'
gem 'font-awesome-sass'
gem 'foreigner'
gem 'gon'
gem 'jbuilder'
gem 'jquery-rails'
gem 'kaminari'
gem 'keen'
gem 'koala'
gem 'meta-tags'
gem 'nokogiri'
gem 'omniauth-facebook'
gem 'omniauth-twitter'
gem 'paper_trail'
gem 'pg'
gem 'rails-i18n'
gem 'ransack'
gem 'react-rails', '~> 1.0.0.pre', github: 'reactjs/react-rails'
gem 'sass-rails'
gem 'sidekiq'
gem 'sidekiq-middleware' # Recommendableで使用
gem 'sinatra', require: nil
gem 'slim'
gem 'recommendable' # gem 'sidekiq' より下に置く必要があるらしい
gem 'twitter'
gem 'uglifier'
gem 'unicorn'
gem 'whenever', require: false

gem 'rails-assets-chartjs'
gem 'rails-assets-EventEmitter.js'
gem 'rails-assets-flux'
gem 'rails-assets-moment'
gem 'rails-assets-uri.js'

group :development, :test do
  gem 'awesome_print'
  gem 'hirb-unicode'
  gem 'hirb'
  gem 'pry-byebug'
  gem 'pry-coolline'
  gem 'pry-rails'
  gem 'rails-flog', require: 'flog'
  gem 'rspec-rails'
end

group :development do
  gem 'aws-sdk'
  gem 'bullet'
  gem 'letter_opener_web'
  gem 'quiet_assets'
  gem 'spring-commands-rspec', require: false
  gem 'spring'
  gem 'thin'
  gem 'web-console'
end

group :production do
  gem 'bugsnag'
  gem 'rails_12factor'
  gem 'skylight'
end

group :test do
  gem 'capybara'
  gem 'coveralls', require: false
  gem 'database_rewinder'
  gem 'factory_girl_rails'
  gem 'poltergeist'
end
