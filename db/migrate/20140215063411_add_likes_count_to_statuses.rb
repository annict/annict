class AddLikesCountToStatuses < ActiveRecord::Migration
  def change
    add_column :statuses, :likes_count, :integer, null: false, default: 0, after: :latest
  end
end
