# frozen_string_literal: true

class AddRatingsCountToWorks < ActiveRecord::Migration[5.1]
  def change
    add_column :works, :ratings_count, :integer, null: false, default: 0
    add_index :works, :ratings_count

    add_column :works, :satisfaction_rate, :float
    add_index :works, :satisfaction_rate
    add_index :works, %i[satisfaction_rate ratings_count]
  end
end
