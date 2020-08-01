# frozen_string_literal: true

class DestroyChannelWorksJob < ApplicationJob
  queue_as :default

  def perform(user, channel)
    return if user.blank? || channel.blank?

    user.works.wanna_watch_and_watching.each do |work|
      channel_work = user.channel_works.find_by(anime_id: work.id, channel_id: channel.id)
      channel_work.destroy if channel_work.present?
    end
  end
end
