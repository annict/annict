class AddImageForPaperclipToProfiles < ActiveRecord::Migration
  def change
    add_attachment :profiles, :tombo_avatar
    add_attachment :profiles, :tombo_background_image
  end
end
