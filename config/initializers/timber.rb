# frozen_string_literal: true

if Rails.env.production?
  config = Timber::Config.instance
  config.logrageify!
  config.integrations.action_view.silence = true
  config.integrations.active_record.silence = true
  config.integrations.rack.http_events.collapse_into_single_event = true
  # Override user context data to not send user email to Timber.
  # https://docs.timber.io/concepts/log-event-json-schema/context/user/
  # https://github.com/timberio/timber-ruby/tree/e9d4582c20e974ff00a039e6ec697784fa1b6ecd#configuration
  config.integrations.rack.user_context.custom_user_hash = lambda do |rack_env|
    user = rack_env["warden"]&.user
    return nil unless user
    {
      id: user.id,
      meta: {
        username: user.username
      }
    }
  end
end
