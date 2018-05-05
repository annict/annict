# frozen_string_literal: true
# == Schema Information
#
# Table name: records
#
#  id         :integer          not null, primary key
#  user_id    :integer          not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_records_on_user_id  (user_id)
#

class Record < ApplicationRecord
  RATING_STATES = %i(bad average good great).freeze

  belongs_to :user
  has_one :episode_record
  has_one :work_record
end
