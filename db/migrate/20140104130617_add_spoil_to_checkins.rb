class AddSpoilToCheckins < ActiveRecord::Migration[4.2]
  def change
    add_column :checkins, :spoil, :boolean, null: false, default: false, after: :comment
  end
end
