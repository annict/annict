# frozen_string_literal: true

class DoNotAllowRecordIdNullOnEpisodeAndWorkRecords < ActiveRecord::Migration[5.2]
  def change
    change_column_null :episode_records, :record_id, false
    change_column_null :work_records, :record_id, false
  end
end
