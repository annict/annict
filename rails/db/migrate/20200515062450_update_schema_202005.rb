# frozen_string_literal: true

class UpdateSchema202005 < ActiveRecord::Migration[6.0]
  def change
    change_column_null :activities, :recipient_id, true
    change_column_null :activities, :recipient_type, true
    change_column_null :activities, :action, true

    add_column :settings, :timeline_mode, :string, null: false, default: "following"
  end
end
