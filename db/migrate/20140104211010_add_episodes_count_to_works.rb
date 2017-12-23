class AddEpisodesCountToWorks < ActiveRecord::Migration[4.2]
  def change
    add_column :works, :episodes_count, :integer, null: false, default: 0, after: :wikipedia_url
    add_index  :works, :episodes_count
  end
end
