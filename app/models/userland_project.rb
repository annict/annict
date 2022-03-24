# frozen_string_literal: true

# == Schema Information
#
# Table name: userland_projects
#
#  id                   :bigint           not null, primary key
#  available            :boolean          default(FALSE), not null
#  description          :text             not null
#  icon_content_type    :string
#  icon_file_name       :string
#  icon_file_size       :integer
#  icon_updated_at      :datetime
#  image_data           :text
#  locale               :string           default("other"), not null
#  name                 :string           not null
#  summary              :string           not null
#  url                  :string           not null
#  created_at           :datetime         not null
#  updated_at           :datetime         not null
#  userland_category_id :bigint           not null
#
# Indexes
#
#  index_userland_projects_on_locale                (locale)
#  index_userland_projects_on_userland_category_id  (userland_category_id)
#
# Foreign Keys
#
#  fk_rails_...  (userland_category_id => userland_categories.id)
#

class UserlandProject < ApplicationRecord
  include UserlandProjectImageUploader::Attachment.new(:image)
  include UgcLocalizable
  include ImageUploadable

  counter_culture :userland_category

  belongs_to :userland_category
  has_many :userland_project_members, dependent: :destroy
  has_many :users, through: :userland_project_members

  validates :description, presence: true
  validates :name, presence: true, length: {maximum: 50}
  validates :summary, presence: true, length: {maximum: 150}
  validates :url, presence: true, url: true

  def image_aspect_ratio(field)
    case field
    when :image
      "1:1"
    else
      raise Annict::Errors::UnknownImageFieldError, "Unexpected field name: #{field}"
    end
  end
end
