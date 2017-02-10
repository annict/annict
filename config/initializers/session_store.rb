# frozen_string_literal: true
# Be sure to restart your server when you modify this file.

Annict::Application.config.session_store :active_record_store,
  expire_after: 10.days,
  domain: :all
