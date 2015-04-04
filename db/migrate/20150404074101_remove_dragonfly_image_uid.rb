class RemoveDragonflyImageUid < ActiveRecord::Migration
  def change
    remove_column :items, :image_uid
    remove_column :profiles, :avatar_uid
    remove_column :profiles, :background_image_uid
  end
end
