# frozen_string_literal: true

if Rails.env.production?
  Resend.api_key = ENV["ANNICT_RESEND_API_KEY"]
end
