source "https://rubygems.org"

ruby "2.2.2"

gem "rails", "4.2.3"

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
gem "browser"
gem "by_star"
gem "coffee-rails"
gem "delayed_job_active_record"
gem "devise"
gem "draper"
gem "email_validator"
gem "enumerize"
gem "figaro"
gem "flutie"
gem "font-awesome-sass"
gem "gon"
gem "hashdiff"
gem "jbuilder"
gem "jquery-rails"
gem "kaminari"
gem "keen"
gem "koala"
gem "meta-tags"
gem "newrelic_rpm"
gem "ngannotate-rails"
gem "nokogiri"
gem "omniauth-facebook"
gem "omniauth-twitter"
gem "paper_trail"
gem "paperclip"
gem "pg"
gem "puma"
gem "pundit"
gem "rails-i18n"
gem "ransack"
gem "rmagick"
gem "sass-rails"
gem "sinatra", require: nil
gem "slim"
gem "traceroute"
gem "twitter"
gem "uglifier"
gem "validate_url"
gem "virtus"

source "https://rails-assets.org" do
  gem "rails-assets-angularjs"
  gem "rails-assets-angular-animate"
  gem "rails-assets-angular-sanitize"
  gem "rails-assets-angulartics"
  gem "rails-assets-chartjs"
  gem "rails-assets-jquery-easing-original"
  gem "rails-assets-lodash"
  gem "rails-assets-moment"
  gem "rails-assets-ng-infinite-scroller-origin"
end

group :development, :test do
  gem "awesome_print"
  gem "did_you_mean", github: "yuki24/did_you_mean"
  gem "activerecord-correctable", github: "yuki24/activerecord-correctable" # did_you_meanより後に配置する必要がある
  gem "hirb-unicode"
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
  gem "spring-commands-rspec", require: false
  gem "spring"
  gem "thin"
end

group :production do
  gem "bugsnag"
  gem "rails_12factor"
end

group :test do
  gem "capybara"
  gem "codeclimate-test-reporter", require: nil
  gem "database_rewinder"
  gem "factory_girl_rails"
  gem "poltergeist"
end
