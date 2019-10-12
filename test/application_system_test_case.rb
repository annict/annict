# frozen_string_literal: true

require "test_helper"

ImageHelper.module_eval do
  def ann_image_url(*)
    "#{ENV.fetch('ANNICT_URL')}/dummy_image"
  end
end

def browser
  return :chrome if ENV["CI"] || ENV["NO_HEADLESS"]

  :headless_chrome
end

def options
  return {} unless ENV["CI"]

  {
    desired_capabilities: Selenium::WebDriver::Remote::Capabilities.chrome(
      chrome_options: {
        args: %w(
          headless
          no-sandbox
          window-size=1280x800
        )
      }
    )
  }
end

class ApplicationSystemTestCase < ActionDispatch::SystemTestCase
  driven_by :selenium, using: browser, options: options
end
