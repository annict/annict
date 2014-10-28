class AddBackgroundImageToProfiles < ActiveRecord::Migration
  def change
    add_column :profiles, :background_image_uid, :string, after: :avatar_uid
  end
end
