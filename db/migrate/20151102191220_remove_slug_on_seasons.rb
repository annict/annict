class RemoveSlugOnSeasons < ActiveRecord::Migration
  def change
    remove_column :seasons, :slug
    change_column_null :seasons, :year, false
  end
end
