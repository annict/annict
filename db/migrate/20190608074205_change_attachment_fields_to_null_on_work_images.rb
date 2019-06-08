# frozen_string_literal: true

class ChangeAttachmentFieldsToNullOnWorkImages < ActiveRecord::Migration[5.2]
  def change
    change_column_null :work_images, :attachment_file_name, true
    change_column_null :work_images, :attachment_file_size, true
    change_column_null :work_images, :attachment_content_type, true
    change_column_null :work_images, :attachment_updated_at, true
    change_column_null :work_images, :image_data, false
  end
end
