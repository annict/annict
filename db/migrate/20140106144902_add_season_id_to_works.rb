class AddSeasonIdToWorks < ActiveRecord::Migration
  def change
    add_column :works, :season_id, :integer, after: :id
    add_foreign_key :works, :seasons
  end
end
