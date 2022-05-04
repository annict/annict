# frozen_string_literal: true

# == Schema Information
#
# Table name: programs
#
#  id                                 :bigint           not null, primary key
#  aasm_state                         :string           default("published"), not null
#  deleted_at                         :datetime
#  minimum_episode_generatable_number :integer          default(1), not null
#  rebroadcast                        :boolean          default(FALSE), not null
#  started_at                         :datetime
#  unpublished_at                     :datetime
#  url                                :string
#  vod_title_code                     :string           default(""), not null
#  vod_title_name                     :string           default(""), not null
#  created_at                         :datetime         not null
#  updated_at                         :datetime         not null
#  channel_id                         :bigint           not null
#  work_id                            :bigint           not null
#
# Indexes
#
#  index_programs_on_channel_id      (channel_id)
#  index_programs_on_deleted_at      (deleted_at)
#  index_programs_on_unpublished_at  (unpublished_at)
#  index_programs_on_vod_title_code  (vod_title_code)
#  index_programs_on_work_id         (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (channel_id => channels.id)
#  fk_rails_...  (work_id => works.id)
#

class Program < ApplicationRecord
  include DbActivityMethods
  include Unpublishable

  DIFF_FIELDS = %i[channel_id work_id url started_at].freeze

  attr_accessor :time_zone

  validates :url, url: {allow_blank: true}

  belongs_to :channel
  belongs_to :work, touch: true
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  has_many :slots, dependent: :destroy

  scope :in_vod, -> { joins(:channel).where(channels: {vod: true}) }

  delegate :name, to: :channel, prefix: true

  before_save :calc_for_timezone

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) { |field, hash|
      hash[field] = send(field)
      hash
    }

    data.delete_if { |_, v| v.blank? }
  end

  def support_en?
    false
  end

  def vod_title_url
    return "" if vod_title_code.blank?

    case channel_id
    when Channel::AMAZON_VIDEO_ID
      "https://www.amazon.co.jp/dp/#{vod_title_code}"
    when Channel::BANDAI_CHANNEL_ID
      "http://www.b-ch.com/ttl/index.php?ttl_c=#{vod_title_code}"
    when Channel::D_ANIME_STORE_ID
      "https://animestore.docomo.ne.jp/animestore/ci_pc?workId=#{vod_title_code}"
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

  def slot_by_episode(episode)
    return unless episode

    slots.only_kept.find_by(episode: episode)
  end

  private

  def calc_for_timezone
    return if time_zone.blank?
    return if started_at.nil?

    started_at = ActiveSupport::TimeZone.new(time_zone).local_to_utc(self.started_at)
    self.started_at = started_at
  end
end
