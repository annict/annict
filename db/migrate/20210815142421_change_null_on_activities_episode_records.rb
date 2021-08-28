# frozen_string_literal: true

class ChangeNullOnActivitiesEpisodeRecords < ActiveRecord::Migration[6.1]
  def change
    change_column_null :activities, :trackable_id, true
    change_column_null :activities, :trackable_type, true

    change_column_null :episode_records, :user_id, true
    change_column_null :episode_records, :episode_id, true
    change_column_null :episode_records, :work_id, true
    change_column_null :episode_records, :record_id, true
  end
end
