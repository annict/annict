# typed: false
# frozen_string_literal: true

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
