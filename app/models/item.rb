# == Schema Information
#
# Table name: items
#
#  id                       :integer          not null, primary key
#  work_id                  :integer
#  name                     :string           not null
#  url                      :string           not null
#  main                     :boolean          default(FALSE), not null
#  created_at               :datetime
#  updated_at               :datetime
#  tombo_image_file_name    :string
#  tombo_image_content_type :string
#  tombo_image_file_size    :integer
#  tombo_image_updated_at   :datetime
#

class Item < ActiveRecord::Base
  has_attached_file :tombo_image

  belongs_to :work

  validates :tombo_image, attachment_presence: true,
                          attachment_content_type: { content_type: /\Aimage/ }
end
