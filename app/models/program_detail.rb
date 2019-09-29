# frozen_string_literal: true
# == Schema Information
#
# Table name: program_details
#
#  id                                 :bigint(8)        not null, primary key
#  channel_id                         :integer          not null
#  work_id                            :integer          not null
#  url                                :string
#  started_at                         :datetime
#  aasm_state                         :string           default("published"), not null
#  created_at                         :datetime         not null
#  updated_at                         :datetime         not null
#  vod_title_code                     :string           default(""), not null
#  vod_title_name                     :string           default(""), not null
#  rebroadcast                        :boolean          default(FALSE), not null
#  minimum_episode_generatable_number :integer          default(1), not null
#
# Indexes
#
#  index_program_details_on_channel_id      (channel_id)
#  index_program_details_on_vod_title_code  (vod_title_code)
#  index_program_details_on_work_id         (work_id)
#

class ProgramDetail < ApplicationRecord
  include AASM
  include DbActivityMethods

  DIFF_FIELDS = %i(channel_id work_id url started_at).freeze

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  attr_accessor :time_zone

  validates :url, url: { allow_blank: true }

  belongs_to :channel
  belongs_to :work, touch: true
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  has_many :programs, dependent: :destroy

  scope :in_vod, -> { joins(:channel).where(channels: { vod: true }) }

  before_save :calc_for_timezone

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end

  def support_en?
    false
  end

  def vod_title_url
    case channel_id
    when Channel::AMAZON_VIDEO_ID
      "https://www.amazon.co.jp/dp/#{vod_title_code}"
    when Channel::BANDAI_CHANNEL_ID
      "http://www.b-ch.com/ttl/index.php?ttl_c=#{vod_title_code}"
    when Channel::D_ANIME_STORE_ID
      "https://anime.dmkt-sp.jp/animestore/ci_pc?workId=#{vod_title_code}"
    when Channel::NICONICO_CHANNEL_ID
      "http://ch.nicovideo.jp/#{vod_title_code}"
    when Channel::NETFLIX_ID
      "https://www.netflix.com/title/#{vod_title_code}"
    when Channel::ABEMA_VIDEO_ID
      "https://abema.tv/video/title/#{vod_title_code}"
    else
      ""
    end
  end

  private

  def calc_for_timezone
    return if time_zone.blank?
    started_at = ActiveSupport::TimeZone.new(time_zone).local_to_utc(self.started_at)
    self.started_at = started_at
  end
end
