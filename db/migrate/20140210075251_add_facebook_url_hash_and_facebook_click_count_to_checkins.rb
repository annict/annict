class AddFacebookUrlHashAndFacebookClickCountToCheckins < ActiveRecord::Migration
  def change
    add_column :checkins, :facebook_url_hash,    :string, after: :twitter_url_hash
    add_column :checkins, :facebook_click_count, :integer, null: false, default: 0, after: :twitter_click_count
    add_index  :checkins, :facebook_url_hash, unique: true
  end
end
