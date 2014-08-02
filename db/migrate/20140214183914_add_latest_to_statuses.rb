class AddLatestToStatuses < ActiveRecord::Migration
  def change
    add_column :statuses, :latest, :boolean, null: false, default: false, after: :kind
    add_index :statuses, [:user_id, :latest]
  end
end
