# frozen_string_literal: true

class AddStatusIdToLatestStatuses < ActiveRecord::Migration[6.0]
  def change
    add_column :latest_statuses, :status_id, :integer

    add_index :latest_statuses, :status_id

    add_foreign_key :latest_statuses, :statuses
  end
end
