class AddCheckinsCountToEpisodes < ActiveRecord::Migration
  def change
    add_column :episodes, :checkins_count, :integer, null: false, default: 0, after: :title
    add_index  :episodes, :checkins_count
  end
end
