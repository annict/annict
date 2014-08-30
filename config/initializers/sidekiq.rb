Sidekiq.configure_server do |config|
  config.redis = { url: "redis://#{REDIS_HOST}:6379/12", namespace: 'sidekiq' }
end
