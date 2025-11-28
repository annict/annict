# frozen_string_literal: true

class CreateInternalStatistics < ActiveRecord::Migration[5.1]
  def change
    create_table :internal_statistics do |t|
      t.string :key, null: false
      t.float :value, null: false
      t.date :date, null: false
      t.timestamps null: false
    end

    add_index :internal_statistics, :key
    add_index :internal_statistics, %i[key date], unique: true
  end
end
