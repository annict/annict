# typed: false
# frozen_string_literal: true

class Slot < ApplicationRecord
  include DbActivityMethods
  include Unpublishable

  DIFF_FIELDS = %i[channel_id episode_id started_at rebroadcast].freeze

  attr_accessor :time_zone, :is_started_at_calced, :shift_time_along_with_after_slots

  belongs_to :channel
  belongs_to :episode, optional: true
  belongs_to :program, optional: true
  belongs_to :work, touch: true
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  has_many :library_entries, dependent: :nullify, foreign_key: :next_slot_id, inverse_of: :next_slot

  validates :channel_id, presence: true
  validates :started_at, presence: true

  scope :episode_published, -> { joins(:episode).merge(Episode.only_kept) }
  scope :with_not_deleted_work, -> { joins(:work).merge(Work.only_kept) }
  scope :with_works, ->(works) { joins(:work).merge(works) }

  before_validation :calc_for_timezone

  def broadcasted?(time = Time.now.in_time_zone("Asia/Tokyo"))
    time > started_at.in_time_zone("Asia/Tokyo")
  end

  def calc_for_timezone
    return if time_zone.blank?
    return if is_started_at_calced

    started_at = ActiveSupport::TimeZone.new(time_zone).local_to_utc(self.started_at)

    self.started_at = started_at
    self.sc_last_update = Time.zone.now
    self.is_started_at_calced = true
  end

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) { |field, hash|
      hash[field] = send(field)
      hash
    }

    data.delete_if { |_, v| v.blank? }
  end

  def state
    now = Time.now.in_time_zone("Asia/Tokyo")
    start = started_at.in_time_zone("Asia/Tokyo")

    if now > start
      return :broadcasting if now.between?(start, start + work.duration.minutes)

      :broadcasted
    else
      return :tonight if start.between?(now, now.beginning_of_day + 1.day + 5.hours)

      :unbroadcast
    end
  end

  def save_all_and_create_activity!
    ActiveRecord::Base.transaction do
      old_started_at = started_at_was

      save_and_create_activity!

      if shift_time_along_with_after_slots == "1" && started_at != old_started_at
        program.slots.where("number > ?", number).each do |slot|
          slot.update!(started_at: slot.started_at + (started_at - old_started_at))
        end
      end
    end
  end
end
