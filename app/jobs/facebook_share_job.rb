# frozen_string_literal: true

class FacebookShareJob < ApplicationJob
  queue_as :default

  def perform(user_id, record_id)
    user = User.find(user_id)
    record = user.records.find(record_id)

    FacebookService.new(user).share!(record)
  end
end
