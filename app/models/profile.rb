# frozen_string_literal: true

# == Schema Information
#
# Table name: profiles
#
#  id                                  :bigint           not null, primary key
#  background_image_animated           :boolean          default(FALSE), not null
#  background_image_data               :text
#  description                         :string(510)      default(""), not null
#  image_data                          :text
#  name                                :string(510)      default(""), not null
#  tombo_avatar_content_type           :string
#  tombo_avatar_file_name              :string
#  tombo_avatar_file_size              :integer
#  tombo_avatar_updated_at             :datetime
#  tombo_background_image_content_type :string
#  tombo_background_image_file_name    :string
#  tombo_background_image_file_size    :integer
#  tombo_background_image_updated_at   :datetime
#  url                                 :string
#  created_at                          :datetime
#  updated_at                          :datetime
#  user_id                             :bigint           not null
#
# Indexes
#
#  profiles_user_id_idx  (user_id)
#  profiles_user_id_key  (user_id) UNIQUE
#
# Foreign Keys
#
#  profiles_user_id_fk  (user_id => users.id) ON DELETE => cascade
#

class Profile < ApplicationRecord
  include ProfileImageUploader::Attachment.new(:image)
  include ProfileImageUploader::Attachment.new(:background_image)
  include ImageUploadable

  belongs_to :user, touch: true

  validates :description, length: {maximum: 150}
  validates :name, presence: true
  validates :url, url: {allow_blank: true}

  before_save :check_animated_gif

  def description=(description)
    value = description.present? ? description.truncate(150) : ""
    write_attribute(:description, value)
  end

  private

  def check_animated_gif
    if background_image_data_changed?
      file = uploaded_file(:background_image, size: :original)
      data = Rails.env.test? ? file.to_io : file.url
      image = MiniMagick::Image.open(data)
      self.background_image_animated = (image.frames.length > 1)
    end

    self
  end
end
