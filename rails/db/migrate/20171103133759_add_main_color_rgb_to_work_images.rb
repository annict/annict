# frozen_string_literal: true

class AddMainColorRgbToWorkImages < ActiveRecord::Migration[5.1]
  def change
    add_column :work_images, :color_rgb, :string, null: false, default: "255,255,255"
  end
end
