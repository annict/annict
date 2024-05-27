# typed: false
# frozen_string_literal: true

class SendAnalyticsEventJob < ApplicationJob
  queue_as :low

  def perform(body)
    Annict::Analytics::Event.post("/collect", body: body)
  end
end
