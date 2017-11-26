# frozen_string_literal: true

class CreateRecordActivityJob < ApplicationJob
  queue_as :default

  def perform(user_id, record_id)
    user = User.find(user_id)
    record = user.records.find(record_id)

    Activity.create! do |a|
      a.user = user
      a.recipient = record.episode
      a.trackable = record
      a.action = "create_record"
      a.work = record.episode.work
      a.episode = record.episode
      a.record = record
    end
  end
end
