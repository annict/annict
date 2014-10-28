class SetDefaultFalseToLatestOnStatuses < ActiveRecord::Migration
  def change
    change_column_default :statuses, :latest, :false
    change_column_default :statuses, :likes_count, 0
  end
end
