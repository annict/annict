# frozen_string_literal: true
# Be sure to restart your server when you modify this file.

Annict::Application.config.session_store :redis_session_store,
  key: "_annict_session",
  domain: :all,
  redis: {
    expire_after: 30.days,
    key_prefix: "annict:session:",
    url: "#{ENV.fetch('REDIS_BLUE_URL')}/0"
  }
