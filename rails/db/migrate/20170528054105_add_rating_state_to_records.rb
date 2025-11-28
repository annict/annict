# frozen_string_literal: true

class AddRatingStateToRecords < ActiveRecord::Migration[5.1]
  def change
    add_column :checkins, :rating_state, :string
    add_index :checkins, :rating_state
  end
end
