# frozen_string_literal: true

class RenameVideoServiceToVodOnChannels < ActiveRecord::Migration[5.1]
  def change
    rename_column :channels, :video_service, :vod
  end
end
