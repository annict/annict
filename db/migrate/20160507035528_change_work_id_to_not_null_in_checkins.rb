# frozen_string_literal: true

class ChangeWorkIdToNotNullInCheckins < ActiveRecord::Migration
  def change
    change_column_null :checkins, :work_id, false
  end
end
