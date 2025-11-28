# frozen_string_literal: true

class AddNullFalseToWatchedAtOnRecords < ActiveRecord::Migration[6.1]
  def change
    change_column_null :records, :watched_at, false
  end
end
