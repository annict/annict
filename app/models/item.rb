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

  belongs_to :work, counter_cache: true

  validates :name, presence: true
  validates :url, presence: true, url: true
  validates :tombo_image, attachment_presence: true,
                          attachment_content_type: { content_type: /\Aimage/ }
  validate :amazon_url

  after_save :switch_main_flag

  private

  def amazon_url
    unless /amazon\.co\.jp\z/ === URI.parse(url).host
      errors.add(:url, "にはAmazon.co.jpの商品URLを入力してください。")
    end
  end

  def switch_main_flag
    if main?
      work.items.where.not(id: id).update_all(main: false)
    end
  end
end
