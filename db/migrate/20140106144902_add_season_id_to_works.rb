class AddSeasonIdToWorks < ActiveRecord::Migration[4.2]
  def change
    add_column :works, :season_id, :integer, after: :id
    add_foreign_key :works, :seasons
  end
end
