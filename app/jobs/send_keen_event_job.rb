# frozen_string_literal: true

class SendKeenEventJob < ApplicationJob
  queue_as :low_priority

  def perform(event_collection, properties)
    Keen.publish(event_collection, properties)
  end
end
