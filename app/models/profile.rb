# == Schema Information
#
# Table name: profiles
#
#  id                                  :integer          not null, primary key
#  user_id                             :integer          not null
#  name                                :string(510)      default(""), not null
#  description                         :string(510)      default(""), not null
#  avatar_uid                          :string(510)
#  background_image_uid                :string(510)
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
#  profiles_user_id_idx  (user_id)
#  profiles_user_id_key  (user_id) UNIQUE
#

class Profile < ActiveRecord::Base
  dragonfly_accessor :avatar do
    default 'public/images/no_image.png'
  end
  dragonfly_accessor :background_image

  has_attached_file :tombo_avatar
  has_attached_file :tombo_background_image

  belongs_to :user

  validates :description, length: { maximum: 150 }
  validates :name, presence: true
  validates :tombo_avatar, attachment_presence: true,
                           attachment_content_type: {
                             content_type: /\Aimage/
                           }
  validates :tombo_background_image, attachment_presence: true,
                                     attachment_content_type: {
                                       content_type: /\Aimage/
                                     }

  before_validation :rename_file
  before_save :check_animated_gif


  def description=(description)
    value = description.present? ? description.truncate(150) : ''
    write_attribute(:description, value)
  end

  private

  def rename_file
    avatar.name = random_file_name(avatar) if avatar.present?
    background_image.name = random_file_name(background_image) if background_image.present?
  end

  def random_file_name(file)
    ext = file.name.scan(/\.[a-zA-Z]+$/).first.presence || ''
    SecureRandom.hex(16) + ext
  end

  def check_animated_gif
    if background_image_uid_changed?
      image = Magick::ImageList.new(background_image.file)
      # sceneが0より大きければGifアニメ画像
      # http://stackoverflow.com/questions/27238816/how-to-tell-if-gif-is-animated
      self.background_image_animated = (image.scene > 0)
    end

    self
  end
end
