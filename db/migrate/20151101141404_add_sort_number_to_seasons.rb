class AddSortNumberToSeasons < ActiveRecord::Migration[4.2]
  def change
    add_column :seasons, :sort_number, :integer
    add_index :seasons, :sort_number, unique: true
  end
end
