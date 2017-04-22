# == Schema Information
#
# Table name: channel_works
#
#  id         :integer          not null, primary key
#  user_id    :integer          not null
#  work_id    :integer          not null
#  channel_id :integer          not null
#  created_at :datetime
#  updated_at :datetime
#
# Indexes
#
#  channel_works_channel_id_idx                  (channel_id)
#  channel_works_user_id_idx                     (user_id)
#  channel_works_user_id_work_id_channel_id_key  (user_id,work_id,channel_id) UNIQUE
#  channel_works_work_id_idx                     (work_id)
#

class ChannelWork < ActiveRecord::Base
  belongs_to :channel
  belongs_to :user
  belongs_to :work, touch: true

  def self.current_channel(work)
    channel_work = find_by(work_id: work.id)
    channel_work.channel if channel_work.present?
  end
end
