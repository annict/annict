class AddSharedTwitterAndSharedFacebookToCheckins < ActiveRecord::Migration
  def change
    add_column :checkins, :shared_twitter, :boolean, null: false, default: false
    add_column :checkins, :shared_facebook, :boolean, null: false, default: false
  end
end
