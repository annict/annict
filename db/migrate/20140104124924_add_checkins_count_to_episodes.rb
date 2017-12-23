class AddCheckinsCountToEpisodes < ActiveRecord::Migration[4.2]
  def change
    add_column :episodes, :checkins_count, :integer, null: false, default: 0, after: :title
    add_index  :episodes, :checkins_count
  end
end
