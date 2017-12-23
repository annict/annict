class UpdateSeasons < ActiveRecord::Migration[4.2]
  def change
    change_column_null :seasons, :sort_number, false
    add_column :seasons, :year, :integer

    add_index :seasons, :year
    add_index :seasons, [:year, :name], unique: true
  end
end
