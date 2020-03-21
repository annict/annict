# frozen_string_literal: true

class CreateChannelWorksJob < ApplicationJob
  queue_as :default

  def perform(user, channel)
    return if user.blank? || channel.blank?

    user.works.wanna_watch_and_watching.each do |work|
      conditions =
        !user.channel_works.exists?(work_id: work.id) &&
        work.channels.only_kept.present? &&
        work.channels.only_kept.exists?(id: channel.id)

      user.channel_works.create(work: work, channel: channel) if conditions
    end
  end
end
