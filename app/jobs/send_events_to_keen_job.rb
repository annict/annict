# frozen_string_literal: true

class SendEventsToKeenJob < ApplicationJob
  queue_as :low_priority

  def perform(collection, properties)
    Keen.publish(collection, properties)
  end
end
