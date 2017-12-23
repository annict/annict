class RemoveSlugOnSeasons < ActiveRecord::Migration[4.2]
  def change
    remove_column :seasons, :slug
    change_column_null :seasons, :year, false
  end
end
