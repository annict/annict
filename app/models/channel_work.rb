# typed: false
# frozen_string_literal: true

class ChannelWork < ApplicationRecord
  belongs_to :channel
  belongs_to :user
  belongs_to :work

  def self.current_channel(work)
    channel_work = find_by(work_id: work.id)
    channel_work.channel if channel_work.present?
  end
end
