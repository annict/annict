# typed: false
# frozen_string_literal: true

class Channel < ApplicationRecord
  include Unpublishable

  AMAZON_VIDEO_ID = 243
  BANDAI_CHANNEL_ID = 107
  D_ANIME_STORE_ID = 241
  D_ANIME_STORE_NICONICO_ID = 306
  NICONICO_CHANNEL_ID = 165
  NETFLIX_ID = 244
  ABEMA_VIDEO_ID = 260

  belongs_to :channel_group
  has_many :programs, dependent: :destroy
  has_many :slots, dependent: :destroy

  scope :with_vod, -> { where(vod: true) }
  scope :without_vod, -> { where.not(vod: true) }

  validates :name, presence: true
  validates :channel_group, presence: true

  def self.fastest(work)
    receivable_channel_ids = pluck(:id)

    if receivable_channel_ids.present? && work.episodes.present?
      conditions = {channel_id: receivable_channel_ids, episode: work.episodes.first}
      fastest_slot = Slot.where(conditions).order(:started_at).first

      fastest_slot.present? ? fastest_slot.channel : nil
    end
  end

  def amazon_video?
    id == AMAZON_VIDEO_ID
  end
end
