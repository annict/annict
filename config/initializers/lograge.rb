# typed: false
# frozen_string_literal: true

if ENV["ANNICT_LOGRAGE_ENABLED"] == "true"
  Rails.application.configure do
    config.lograge.enabled = true

    config.lograge.custom_payload do |controller|
      controller.respond_to?(:lograge_payload) ? controller.lograge_payload : {}
    end
  end
end
