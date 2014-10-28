class AddDefaultTrueToPublishedOnChannels < ActiveRecord::Migration
  def change
    change_column_default :channels, :published, true
  end
end
