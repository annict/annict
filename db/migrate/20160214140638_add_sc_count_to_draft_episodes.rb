class AddScCountToDraftEpisodes < ActiveRecord::Migration[4.2]
  def change
    add_column :draft_episodes, :sc_count, :integer
  end
end
