# frozen_string_literal: true
# Be sure to restart your server when you modify this file.

Annict::Application.config.session_store :redis_store,
  servers: ["#{ENV.fetch('REDIS_BLUE_URL')}/session"],
  expire_after: 30.days,
  domain: :all,
  key: "_ann_session"
