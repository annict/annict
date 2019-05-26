# frozen_string_literal: true
# == Schema Information
#
# Table name: userland_projects
#
#  id                   :integer          not null, primary key
#  userland_category_id :integer          not null
#  name                 :string           not null
#  summary              :string           not null
#  description          :text             not null
#  url                  :string           not null
#  icon_file_name       :string
#  icon_content_type    :string
#  icon_file_size       :integer
#  icon_updated_at      :datetime
#  available            :boolean          default(FALSE), not null
#  created_at           :datetime         not null
#  updated_at           :datetime         not null
#  locale               :string           default("other"), not null
#  image_data           :text
#
# Indexes
#
#  index_userland_projects_on_locale                (locale)
#  index_userland_projects_on_userland_category_id  (userland_category_id)
#

class UserlandProject < ApplicationRecord
  include UserlandProjectImageUploader::Attachment.new(:image)
  include LocaleDetectable
  include ImageUploadable

  belongs_to :userland_category, counter_cache: true
  has_many :userland_project_members, dependent: :destroy
  has_many :users, through: :userland_project_members

  validates :description, presence: true
  validates :name, presence: true, length: { maximum: 50 }
  validates :summary, presence: true, length: { maximum: 150 }
  validates :url, presence: true, url: true
end
