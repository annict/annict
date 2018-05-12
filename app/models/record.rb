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
#
# Indexes
#
#  index_records_on_user_id  (user_id)
#

class Record < ApplicationRecord
  RATING_STATES = %i(bad average good great).freeze

  is_impressionable counter_cache: true, unique: true

  belongs_to :user
  has_one :episode_record
  has_one :work_record

  def episode_record?
    episode_record.present?
  end
end
