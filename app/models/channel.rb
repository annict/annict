# frozen_string_literal: true
# == Schema Information
#
# Table name: channels
#
#  id               :integer          not null, primary key
#  aasm_state       :string           default("published"), not null
#  deleted_at       :datetime
#  name             :string           not null
#  sc_chid          :integer
#  sort_number      :integer          default(0), not null
#  vod              :boolean          default(FALSE)
#  created_at       :datetime
#  updated_at       :datetime
#  channel_group_id :integer          not null
#
# Indexes
#
#  channels_channel_group_id_idx  (channel_group_id)
#  channels_sc_chid_key           (sc_chid) UNIQUE
#  index_channels_on_deleted_at   (deleted_at)
#  index_channels_on_sort_number  (sort_number)
#  index_channels_on_vod          (vod)
#
# Foreign Keys
#
#  channels_channel_group_id_fk  (channel_group_id => channel_groups.id) ON DELETE => cascade
#

class Channel < ApplicationRecord
  include SoftDeletable

  AMAZON_VIDEO_ID = 243
  BANDAI_CHANNEL_ID = 107
  D_ANIME_STORE_ID = 241
  NICONICO_CHANNEL_ID = 165
  NETFLIX_ID = 244
  ABEMA_VIDEO_ID = 260

  belongs_to :channel_group
  has_many :programs, dependent: :destroy
  has_many :slots, dependent: :destroy

  scope :with_vod, -> { where(vod: true) }

  validates :name, presence: true
  validates :channel_group, presence: true

  def self.fastest(work)
    receivable_channel_ids = pluck(:id)

    if receivable_channel_ids.present? && work.episodes.present?
      conditions = { channel_id: receivable_channel_ids, episode: work.episodes.first }
      fastest_slot = Slot.where(conditions).order(:started_at).first

      fastest_slot.present? ? fastest_slot.channel : nil
    end
  end

  def soft_delete_with_children
    soft_delete
    programs.without_deleted.each(&:soft_delete)
  end

  def amazon_video?
    id == AMAZON_VIDEO_ID
  end
end
