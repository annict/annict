class AddImageForPaperclipToItems < ActiveRecord::Migration[4.2]
  def change
    add_attachment :items, :tombo_image
  end
end
