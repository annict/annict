class AddWatchersCountToWorks < ActiveRecord::Migration[4.2]
  def change
    add_column :works, :watchers_count, :integer, null: false, default: 0, after: :episodes_count
    add_index  :works, :watchers_count
  end
end
