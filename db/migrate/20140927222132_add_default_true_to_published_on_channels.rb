class AddDefaultTrueToPublishedOnChannels < ActiveRecord::Migration[4.2]
  def change
    change_column_default :channels, :published, true
  end
end
