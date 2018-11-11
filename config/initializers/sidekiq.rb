# frozen_string_literal: true

Sidekiq.configure_server do |config|
  config.redis = { url: "#{ENV.fetch('REDIS_BLUE_URL')}/sidekiq" }
end

Sidekiq.configure_client do |config|
  config.redis = { url: "#{ENV.fetch('REDIS_BLUE_URL')}/sidekiq" }
end
