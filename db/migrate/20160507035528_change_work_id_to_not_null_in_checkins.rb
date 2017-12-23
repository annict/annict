# frozen_string_literal: true

class ChangeWorkIdToNotNullInCheckins < ActiveRecord::Migration[4.2]
  def change
    change_column_null :checkins, :work_id, false
  end
end
