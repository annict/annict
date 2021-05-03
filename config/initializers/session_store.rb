# frozen_string_literal: true

# Be sure to restart your server when you modify this file.

Annict::Application.config.session_store :active_record_store,
  key: "_annict_session_v201904",
  domain: :all,
  expire_after: 30.days
