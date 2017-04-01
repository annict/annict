# == Schema Information
#
# Table name: items
#
#  id                       :integer          not null, primary key
#  work_id                  :integer
#  name                     :string           not null
#  url                      :string           not null
#  created_at               :datetime
#  updated_at               :datetime
#  tombo_image_file_name    :string
#  tombo_image_content_type :string
#  tombo_image_file_size    :integer
#  tombo_image_updated_at   :datetime
#
# Indexes
#
#  index_items_on_work_id  (work_id) UNIQUE
#

class Item < ActiveRecord::Base
  include DbActivityMethods
  include ItemCommon

  belongs_to :work, touch: true
  has_many :draft_items, dependent: :destroy
end
