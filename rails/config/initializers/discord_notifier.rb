# typed: false
# frozen_string_literal: true

require "discord_notifier"

Discord::Notifier.setup do |config|
  config.url = ""
  config.username = "Notifier"
  config.avatar_url = ""

  # Defaults to `false`
  config.wait = true
end
