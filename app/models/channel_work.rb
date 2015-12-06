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
#  index_channel_works_on_user_id_and_work_id                 (user_id,work_id)
#  index_channel_works_on_user_id_and_work_id_and_channel_id  (user_id,work_id,channel_id) UNIQUE
#

class ChannelWork < ActiveRecord::Base
  belongs_to :channel
  belongs_to :user
  belongs_to :work

  def self.current_channel(work)
    channel_work = find_by(work_id: work.id)
    channel_work.channel if channel_work.present?
  end
end
