# typed: false
# frozen_string_literal: true

Rails.application.config.middleware.insert_before(
  Rack::Runtime,
  Rack::Timeout,
  service_timeout: ENV.fetch("RACK_TIMEOUT_SERVICE_TIMEOUT", 15).to_i,
  term_on_timeout: ENV.fetch("RACK_TIMEOUT_TERM_ON_TIMEOUT", 1).to_i
)
