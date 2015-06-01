# == Schema Information
#
# Table name: edit_request_images
#
#  id                 :integer          not null, primary key
#  edit_request_id    :integer          not null
#  image_file_name    :string           not null
#  image_content_type :string           not null
#  image_file_size    :integer          not null
#  image_updated_at   :datetime         not null
#  created_at         :datetime         not null
#  updated_at         :datetime         not null
#
# Indexes
#
#  index_edit_request_images_on_edit_request_id  (edit_request_id)
#

class EditRequestImage < ActiveRecord::Base
  has_attached_file :image

  validates :image, attachment_presence: true,
                   attachment_content_type: { content_type: /\Aimage/ }
end
