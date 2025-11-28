# typed: false
# frozen_string_literal: true

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

    VodHelper.vod_title_url(channel_id, vod_title_code) or ""
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
