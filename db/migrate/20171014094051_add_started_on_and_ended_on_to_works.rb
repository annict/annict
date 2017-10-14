# frozen_string_literal: true

class AddStartedOnAndEndedOnToWorks < ActiveRecord::Migration[5.1]
  def change
    add_column :works, :started_on, :date
    add_column :works, :ended_on, :date
  end
end
