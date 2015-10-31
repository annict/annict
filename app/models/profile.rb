# == Schema Information
#
# Table name: profiles
#
#  id                                  :integer          not null, primary key
#  user_id                             :integer          not null
#  name                                :string           default(""), not null
#  description                         :string           default(""), not null
#  created_at                          :datetime
#  updated_at                          :datetime
#  background_image_animated           :boolean          default(FALSE), not null
#  tombo_avatar_file_name              :string
#  tombo_avatar_content_type           :string
#  tombo_avatar_file_size              :integer
#  tombo_avatar_updated_at             :datetime
#  tombo_background_image_file_name    :string
#  tombo_background_image_content_type :string
#  tombo_background_image_file_size    :integer
#  tombo_background_image_updated_at   :datetime
#
# Indexes
#
#  index_profiles_on_user_id  (user_id) UNIQUE
#

class Profile < ActiveRecord::Base
  has_attached_file :tombo_avatar
  has_attached_file :tombo_background_image

  belongs_to :user

  validates :description, length: { maximum: 150 }
  validates :name, presence: true
  validates :tombo_avatar, attachment_content_type: {
                             content_type: /\Aimage/
                           }
  validates :tombo_background_image, attachment_content_type: {
                                       content_type: /\Aimage/
                                     }

  before_save :check_animated_gif


  def description=(description)
    value = description.present? ? description.truncate(150) : ''
    write_attribute(:description, value)
  end

  private

  def check_animated_gif
    if tombo_background_image_updated_at_changed?
      file_path = Paperclip.io_adapters.for(tombo_background_image).path
      image = Magick::ImageList.new(file_path)
      # sceneが0より大きければGifアニメ画像
      # http://stackoverflow.com/questions/27238816/how-to-tell-if-gif-is-animated
      self.background_image_animated = (image.scene > 0)
    end

    self
  end
end
