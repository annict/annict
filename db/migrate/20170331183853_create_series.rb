# frozen_string_literal: true

class CreateSeries < ActiveRecord::Migration[5.0]
  def change
    create_table :series do |t|
      t.string :name, null: false
      t.string :name_ro, null: false, default: ""
      t.string :name_en, null: false, default: ""
      t.string :aasm_state, null: false, default: "published"
      t.integer :series_works_count, null: false, default: 0
      t.timestamps null: false
    end

    add_presence_constraint :series, :name

    add_inclusion_constraint :series, :aasm_state, in: %w(published hidden)
  end
end
