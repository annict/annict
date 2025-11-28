# typed: false
# frozen_string_literal: true

require "capybara/rspec"
require "capybara-playwright-driver"

Capybara.register_driver(:playwright) do |app|
  Capybara::Playwright::Driver.new(app,
    browser_type: ENV.fetch("ANNICT_CAPYBARA_BROWSER", "chromium").to_sym,
    headless: ENV.fetch("ANNICT_CAPYBARA_HEADLESS", "true") == "true")
end

Capybara.configure do |config|
  config.default_driver = :rack_test
  config.javascript_driver = :playwright
  config.server = :puma
  config.server_host = "0.0.0.0"
  config.server_port = 4000
  config.app_host = ENV.fetch("ANNICT_URL")
  config.default_max_wait_time = 5
  config.disable_animation = true
end

RSpec.configure do |config|
  config.before(:each, type: :system) do
    driven_by :rack_test
  end

  config.before(:each, type: :system, js: true) do
    driven_by :playwright
  end
end
