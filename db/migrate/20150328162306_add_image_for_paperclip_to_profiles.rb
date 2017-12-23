class AddImageForPaperclipToProfiles < ActiveRecord::Migration[4.2]
  def change
    add_attachment :profiles, :tombo_avatar
    add_attachment :profiles, :tombo_background_image
  end
end
