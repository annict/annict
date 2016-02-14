class AddScCountToDraftEpisodes < ActiveRecord::Migration
  def change
    add_column :draft_episodes, :sc_count, :integer
  end
end
