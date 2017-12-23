class AddBackgroundImageAnimatedToProfiles < ActiveRecord::Migration[4.2]
  def change
    add_column :profiles, :background_image_animated, :boolean,
               null: false, default: false
  end
end
