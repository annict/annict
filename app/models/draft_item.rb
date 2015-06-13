# == Schema Information
#
# Table name: draft_items
#
#  id                       :integer          not null, primary key
#  item_id                  :integer
#  work_id                  :integer          not null
#  name                     :string           not null
#  url                      :string           not null
#  main                     :boolean          default(FALSE), not null
#  tombo_image_file_name    :string           not null
#  tombo_image_content_type :string           not null
#  tombo_image_file_size    :integer          not null
#  tombo_image_updated_at   :datetime         not null
#  created_at               :datetime         not null
#  updated_at               :datetime         not null
#
# Indexes
#
#  index_draft_items_on_item_id  (item_id)
#  index_draft_items_on_work_id  (work_id)
#

class DraftItem < ActiveRecord::Base
  include ItemCommon

  belongs_to :origin, class_name: "Item", foreign_key: :item_id
  belongs_to :work
  has_one :edit_request, as: :draft_resource

  accepts_nested_attributes_for :edit_request
end
