# frozen_string_literal: true

class ProgramDetail < ApplicationRecord
  extend Enumerize
  include AASM
  include DbActivityMethods

  DIFF_FIELDS = %i(channel_id work_id url started_at repeat_on).freeze

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  enumerize :repeat_on, in: %w(weekly daily)

  validates :url, url: { allow_blank: true }

  belongs_to :channel
  belongs_to :work
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
end
