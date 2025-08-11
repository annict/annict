# typed: false
# frozen_string_literal: true

Sentry.init do |config|
  config.dsn = ENV.fetch("SENTRY_DSN")
  config.breadcrumbs_logger = %i[active_support_logger http_logger]

  # Set tracesSampleRate to 1.0 to capture 100%
  # of transactions for performance monitoring.
  # We recommend adjusting this value in production
  config.traces_sample_rate = 0.5

  filter = ActiveSupport::ParameterFilter.new(Rails.application.config.filter_parameters)
  config.before_send = lambda do |event, hint|
    # Use Rails' parameter filter to sanitize the event
    filter.filter(event.to_hash)
  end
end
