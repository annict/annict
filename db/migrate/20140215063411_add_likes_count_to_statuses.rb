class AddLikesCountToStatuses < ActiveRecord::Migration[4.2]
  def change
    add_column :statuses, :likes_count, :integer, null: false, default: 0, after: :latest
  end
end
