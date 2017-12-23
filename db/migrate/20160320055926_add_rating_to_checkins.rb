# frozen_string_literal: true

class AddRatingToCheckins < ActiveRecord::Migration[4.2]
  def change
    add_column :checkins, :rating, :float
  end
end
