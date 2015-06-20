class AddDiffAndValueFieldsToEditRequests < ActiveRecord::Migration
  def change
    add_column :edit_requests, :diffs, :json, null: false
    add_column :edit_requests, :origin_values, :json
    add_column :edit_requests, :draft_values, :json, null: false
  end
end
