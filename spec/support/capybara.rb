# typed: false
# frozen_string_literal: true

require "capybara/rspec"
require "capybara-playwright-driver"

Capybara.register_driver(:playwright) do |app|
  Capybara::Playwright::Driver.new(app,
    browser_type: ENV.fetch("ANNICT_CAPYBARA_BROWSER", "chromium").to_sym,
    headless: ENV.fetch("ANNICT_CAPYBARA_HEADLESS", "true") == "true")
end

Capybara.javascript_driver = :playwright
Capybara.default_driver = :rack_test

RSpec.configure do |config|
  config.before(:each, type: :system) do
    driven_by :rack_test
  end

  config.before(:each, type: :system, js: true) do
    driven_by :playwright
  end
end
