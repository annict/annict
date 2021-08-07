# frozen_string_literal: true

class CreateSeriesWorks < ActiveRecord::Migration[5.0]
  def change
    create_table :series_works do |t|
      t.integer :series_id, null: false
      t.integer :work_id, null: false
      t.string :summary, null: false, default: ""
      t.string :summary_en, null: false, default: ""
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :series_works, :series_id
    add_index :series_works, :work_id
    add_index :series_works, %i[series_id work_id], unique: true

    add_foreign_key :series_works, :series
    add_foreign_key :series_works, :works
  end
end
