# frozen_string_literal: true

# Version of your assets, change this if you want to expire all your assets.
Rails.application.config.assets.version = "1.0"

%w(fonts).each do |dir_name|
  Rails.application.config.assets.paths << "#{Rails.root}/app/assets/#{dir_name}"
end

Rails.application.config.assets.precompile += %w(
  db.scss
  mobile.scss
  mobile.js.coffee
  pc.scss
  pc.js.coffee
)
