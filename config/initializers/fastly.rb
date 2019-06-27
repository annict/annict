# frozen_string_literal: true

FastlyRails.configure do |c|
  # Fastly api key, required
  c.api_key = ENV.fetch("FASTLY_API_KEY")
  # time in seconds, optional, defaults to 2592000 (30 days)
  c.max_age = 2_592_000
  # time in seconds, optional, defaults to nil
  c.stale_while_revalidate = 86_400
  # time in seconds, optional, defaults to nil
  c.stale_if_error = 86_400
  # The Fastly service you will be using, required
  c.service_id = ENV.fetch("FASTLY_SERVICE_ID")
  # No need to configure a client locally (AVAILABLE ONLY AS OF 0.4.0)
  c.purging_enabled = !Rails.env.development?
end
