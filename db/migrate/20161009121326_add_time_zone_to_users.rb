# frozen_string_literal: true

class AddTimeZoneToUsers < ActiveRecord::Migration[5.0]
  def change
    add_column :users, :time_zone, :string, null: false, default: ""
    add_column :users, :locale, :string, null: false, default: ""
  end
end
