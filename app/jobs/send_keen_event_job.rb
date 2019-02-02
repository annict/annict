# frozen_string_literal: true

class SendKeenEventJob < ApplicationJob
  queue_as :low

  def perform(action, data)
    Keen.publish(action, data)
  end
end
