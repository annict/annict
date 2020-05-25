# frozen_string_literal: true

class ChangeNullActivityGroupIdOnActivities < ActiveRecord::Migration[6.0]
  def change
    change_column_null :activities, :activity_group_id, false
  end
end
