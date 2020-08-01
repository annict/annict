# frozen_string_literal: true

# == Schema Information
#
# Table name: channel_animes
#
#  id         :bigint           not null, primary key
#  created_at :datetime
#  updated_at :datetime
#  anime_id   :bigint           not null
#  channel_id :bigint           not null
#  user_id    :bigint           not null
#
# Indexes
#
#  channel_works_channel_id_idx                  (channel_id)
#  channel_works_user_id_idx                     (user_id)
#  channel_works_user_id_work_id_channel_id_key  (user_id,anime_id,channel_id) UNIQUE
#  channel_works_work_id_idx                     (anime_id)
#
# Foreign Keys
#
#  channel_works_channel_id_fk  (channel_id => channels.id) ON DELETE => cascade
#  channel_works_user_id_fk     (user_id => users.id) ON DELETE => cascade
#  channel_works_work_id_fk     (anime_id => animes.id) ON DELETE => cascade
#

class ChannelAnime < ApplicationRecord
  belongs_to :channel
  belongs_to :user
  belongs_to :work

  def self.current_channel(work)
    channel_work = find_by(anime_id: work.id)
    channel_work.channel if channel_work.present?
  end
end
