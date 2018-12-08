# frozen_string_literal: true

if Rails.env.production?
  config = Timber::Config.instance
  config.logrageify!
  config.integrations.action_view.silence = true
  config.integrations.active_record.silence = true
  config.integrations.rack.http_events.collapse_into_single_event = true
end
