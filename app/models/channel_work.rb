# frozen_string_literal: true

# == Schema Information
#
# Table name: channel_works
#
#  id         :bigint           not null, primary key
#  created_at :timestamptz
#  updated_at :timestamptz
#  channel_id :bigint           not null
#  user_id    :bigint           not null
#  work_id    :bigint           not null
#
# Indexes
#
#  channel_works_channel_id_idx                  (channel_id)
#  channel_works_user_id_idx                     (user_id)
#  channel_works_user_id_work_id_channel_id_key  (user_id,work_id,channel_id) UNIQUE
#  channel_works_work_id_idx                     (work_id)
#
# Foreign Keys
#
#  channel_works_channel_id_fk  (channel_id => channels.id) ON DELETE => cascade
#  channel_works_user_id_fk     (user_id => users.id) ON DELETE => cascade
#  channel_works_work_id_fk     (work_id => works.id) ON DELETE => cascade
#

class ChannelWork < ApplicationRecord
  belongs_to :channel
  belongs_to :user
  belongs_to :work

  def self.current_channel(work)
    channel_work = find_by(work_id: work.id)
    channel_work.channel if channel_work.present?
  end
end
