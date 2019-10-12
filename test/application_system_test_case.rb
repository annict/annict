# frozen_string_literal: true

require "test_helper"

ImageHelper.module_eval do
  def ann_image_url(*)
    "#{ENV.fetch('ANNICT_URL')}/dummy_image"
  end
end


class ApplicationSystemTestCase < ActionDispatch::SystemTestCase
  capabilities = Selenium::WebDriver::Remote::Capabilities.chrome(
    chrome_options: {
      args: %w(
        headless
        no-sandbox
        window-size=1280x800
      )
    }
  )
  driven_by :selenium,
    using: :chrome,
    options: { desired_capabilities: capabilities }
end
