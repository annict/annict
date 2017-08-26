# frozen_string_literal: true
# == Schema Information
#
# Table name: program_details
#
#  id         :integer          not null, primary key
#  channel_id :integer          not null
#  work_id    :integer          not null
#  url        :string
#  started_at :datetime
#  aasm_state :string           default("published"), not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_program_details_on_channel_id              (channel_id)
#  index_program_details_on_channel_id_and_work_id  (channel_id,work_id) UNIQUE
#  index_program_details_on_work_id                 (work_id)
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
  belongs_to :work
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy

  scope :in_video_service, -> { joins(:channel).where(channels: { video_service: true }) }

  before_save :calc_for_timezone

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end

  private

  def calc_for_timezone
    return if time_zone.blank?
    started_at = ActiveSupport::TimeZone.new(time_zone).local_to_utc(self.started_at)
    self.started_at = started_at
  end
end
