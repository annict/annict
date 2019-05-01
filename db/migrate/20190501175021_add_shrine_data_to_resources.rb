# frozen_string_literal: true

class AddShrineDataToResources < ActiveRecord::Migration[5.2]
  def change
    add_column :work_images, :file_data, :text
  end
end
