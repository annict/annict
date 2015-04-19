# == Schema Information
#
# Table name: items
#
#  id                       :integer          not null, primary key
#  work_id                  :integer
#  name                     :string(510)      not null
#  url                      :string(510)      not null
#  main                     :boolean          not null
#  created_at               :datetime
#  updated_at               :datetime
#  tombo_image_file_name    :string
#  tombo_image_content_type :string
#  tombo_image_file_size    :integer
#  tombo_image_updated_at   :datetime
#
# Indexes
#
#  items_work_id_idx  (work_id)
#

class Item < ActiveRecord::Base
  has_attached_file :tombo_image

  belongs_to :work

  validates :tombo_image, attachment_presence: true,
                          attachment_content_type: { content_type: /\Aimage/ }
end
