# frozen_string_literal: true

class AddYearAndSeasonToWorks < ActiveRecord::Migration[5.0]
  def change
    add_column :works, :season_year, :integer
    add_column :works, :season_name, :integer
    add_index :works, :season_year
    add_index :works, %i[season_year season_name]
  end
end
