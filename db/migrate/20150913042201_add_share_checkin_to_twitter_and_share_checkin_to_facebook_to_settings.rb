class AddShareCheckinToTwitterAndShareCheckinToFacebookToSettings < ActiveRecord::Migration[4.2]
  def change
    add_column :settings, :share_record_to_twitter, :boolean, default: false
    add_column :settings, :share_record_to_facebook, :boolean, default: false
  end
end
