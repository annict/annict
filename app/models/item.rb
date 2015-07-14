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
  include ItemCommon

  has_paper_trail only: DIFF_FIELDS

  belongs_to :work, counter_cache: true
  has_many :draft_items, dependent: :destroy

  before_save :switch_main_flag

  private

  def switch_main_flag
    if main?
      work.items.where.not(id: id).update_all(main: false)
    end
  end
end
