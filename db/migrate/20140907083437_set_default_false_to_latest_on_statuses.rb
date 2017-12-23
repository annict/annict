class SetDefaultFalseToLatestOnStatuses < ActiveRecord::Migration[4.2]
  def change
    change_column_default :statuses, :latest, :false
    change_column_default :statuses, :likes_count, 0
  end
end
