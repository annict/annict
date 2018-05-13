# frozen_string_literal: true
# == Schema Information
#
# Table name: records
#
#  id                :integer          not null, primary key
#  user_id           :integer          not null
#  created_at        :datetime         not null
#  updated_at        :datetime         not null
#  impressions_count :integer          default(0), not null
#  work_id           :integer
#
# Indexes
#
#  index_records_on_user_id  (user_id)
#  index_records_on_work_id  (work_id)
#

class Record < ApplicationRecord
  include AASM

  RATING_STATES = %i(bad average good great).freeze

  is_impressionable counter_cache: true, unique: true

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :user, counter_cache: true
  belongs_to :work, counter_cache: true
  has_one :episode_record, dependent: :destroy
  has_one :work_record, dependent: :destroy

  def episode_record?
    episode_record.present?
  end
end
