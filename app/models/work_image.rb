# frozen_string_literal: true

# == Schema Information
#
# Table name: work_images
#
#  id                      :bigint           not null, primary key
#  asin                    :string           default(""), not null
#  attachment_content_type :string
#  attachment_file_name    :string
#  attachment_file_size    :integer
#  attachment_updated_at   :datetime
#  copyright               :string           default(""), not null
#  image_data              :text             not null
#  created_at              :datetime         not null
#  updated_at              :datetime         not null
#  user_id                 :bigint           not null
#  work_id                 :bigint           not null
#
# Indexes
#
#  index_work_images_on_user_id  (user_id)
#  index_work_images_on_work_id  (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_id => works.id)
#

class WorkImage < ApplicationRecord
  T.unsafe(self).include WorkImageUploader::Attachment.new(:image)
  include ImageUploadable

  self.ignored_columns = %w[color_rgb]

  validates :copyright, presence: true

  belongs_to :work, touch: true
  belongs_to :user
end
