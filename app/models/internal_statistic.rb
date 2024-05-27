# typed: false
# frozen_string_literal: true

# == Schema Information
#
# Table name: internal_statistics
#
#  id         :bigint           not null, primary key
#  date       :date             not null
#  key        :string           not null
#  value      :float            not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_internal_statistics_on_key           (key)
#  index_internal_statistics_on_key_and_date  (key,date) UNIQUE
#

class InternalStatistic < ApplicationRecord
  extend Enumerize

  enumerize :key, in: %i[
    users_count_active_in_all_users_past_week
    users_count_active_in_new_users_past_week
    users_count_created_episode_records_in_all
    users_count_created_episode_records_in_past_week
    users_count_created_episode_records_in_yesterday
    users_count_created_statuses_in_all
    users_count_created_statuses_in_past_week
    users_count_created_statuses_in_yesterday
    users_count_created_work_records_in_all
    users_count_created_work_records_in_past_week
    users_count_created_work_records_in_yesterday
    users_count_registered_in_all
    users_count_registered_in_past_week
    users_count_registered_in_yesterday
  ]
end
