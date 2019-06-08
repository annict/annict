# frozen_string_literal: true
# == Schema Information
#
# Table name: programs
#
#  id                :integer          not null, primary key
#  channel_id        :integer          not null
#  episode_id        :integer
#  work_id           :integer          not null
#  started_at        :datetime         not null
#  sc_last_update    :datetime
#  created_at        :datetime
#  updated_at        :datetime
#  sc_pid            :integer
#  rebroadcast       :boolean          default(FALSE), not null
#  aasm_state        :string           default("published"), not null
#  program_detail_id :integer
#  number            :integer
#  irregular         :boolean          default(FALSE), not null
#
# Indexes
#
#  index_programs_on_aasm_state                        (aasm_state)
#  index_programs_on_program_detail_id                 (program_detail_id)
#  index_programs_on_program_detail_id_and_episode_id  (program_detail_id,episode_id) UNIQUE
#  index_programs_on_program_detail_id_and_number      (program_detail_id,number) UNIQUE
#  index_programs_on_program_detail_id_and_started_at  (program_detail_id,started_at) UNIQUE
#  index_programs_on_sc_pid                            (sc_pid) UNIQUE
#  programs_channel_id_idx                             (channel_id)
#  programs_episode_id_idx                             (episode_id)
#  programs_work_id_idx                                (work_id)
#

class Program < ApplicationRecord
  include AASM
  include DbActivityMethods

  DIFF_FIELDS = %i(channel_id episode_id started_at rebroadcast).freeze

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  attr_accessor :time_zone, :is_started_at_calced

  belongs_to :channel
  belongs_to :episode, optional: true
  belongs_to :program_detail, optional: true
  belongs_to :work
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy

  validates :channel_id, presence: true
  validates :started_at, presence: true, uniqueness: { scope: :program_detail_id }

  scope :episode_published, -> { joins(:episode).merge(Episode.published) }
  scope :work_published, -> { joins(:work).merge(Work.published) }

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
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end

  def state
    now = Time.now.in_time_zone("Asia/Tokyo")
    start = started_at.in_time_zone("Asia/Tokyo")

    if now > start
      return :broadcasting if now.between?(start, start + work.duration.minutes)
      return :broadcasted
    else
      return :tonight if start.between?(now, now.beginning_of_day + 1.day + 5.hours)
      return :unbroadcast
    end
  end
end
