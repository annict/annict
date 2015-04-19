class AddImageForPaperclipToItems < ActiveRecord::Migration
  def change
    add_attachment :items, :tombo_image
  end
end
