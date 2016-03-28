# frozen_string_literal: true

class AddRatingToCheckins < ActiveRecord::Migration
  def change
    add_column :checkins, :rating, :float
  end
end
