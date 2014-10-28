class AddCommentsCountToCheckins < ActiveRecord::Migration
  def change
    add_column :checkins, :comments_count, :integer, null: false, default: 0, after: :facebook_click_count
  end
end
