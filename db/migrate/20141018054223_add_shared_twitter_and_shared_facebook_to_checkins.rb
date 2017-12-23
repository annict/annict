class AddSharedTwitterAndSharedFacebookToCheckins < ActiveRecord::Migration[4.2]
  def change
    add_column :checkins, :shared_twitter, :boolean, null: false, default: false
    add_column :checkins, :shared_facebook, :boolean, null: false, default: false
  end
end
