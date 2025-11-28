# frozen_string_literal: true

class AddShrineDataToResources < ActiveRecord::Migration[5.2]
  def change
    add_column :items, :image_data, :text
    add_column :profiles, :image_data, :text
    add_column :profiles, :background_image_data, :text
    add_column :pvs, :image_data, :text
    add_column :userland_projects, :image_data, :text
    add_column :work_images, :image_data, :text
  end
end
