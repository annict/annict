class AddSpoilToCheckins < ActiveRecord::Migration
  def change
    add_column :checkins, :spoil, :boolean, null: false, default: false, after: :comment
  end
end
