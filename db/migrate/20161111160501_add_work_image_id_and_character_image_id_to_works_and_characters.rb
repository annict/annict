# frozen_string_literal: true

class AddWorkImageIdAndCharacterImageIdToWorksAndCharacters < ActiveRecord::Migration[5.0]
  def change
    add_column :works, :work_image_id, :integer
    add_column :characters, :character_image_id, :integer

    add_index :works, :work_image_id
    add_index :characters, :character_image_id

    add_foreign_key :works, :work_images
    add_foreign_key :characters, :character_images
  end
end
