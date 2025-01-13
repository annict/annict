# typed: false
# frozen_string_literal: true

class UserlandProject < ApplicationRecord
  T.unsafe(self).include UserlandProjectImageUploader::Attachment.new(:image)
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
end
