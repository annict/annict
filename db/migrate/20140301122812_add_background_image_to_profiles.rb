class AddBackgroundImageToProfiles < ActiveRecord::Migration[4.2]
  def change
    add_column :profiles, :background_image_uid, :string, after: :avatar_uid
  end
end
