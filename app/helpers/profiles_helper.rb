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
#  url                                 :string
#
# Indexes
#
#  index_profiles_on_user_id  (user_id) UNIQUE
#

module ProfilesHelper
  def profile_background_image_url(profile, options)
    background_image = profile.tombo_background_image
    field = background_image.present? ? :tombo_background_image : :tombo_avatar
    image = profile.send(field)

    # プロフィール背景画像がGifアニメのときは、S3に保存された画像をそのまま返す
    if background_image.present? && profile.background_image_animated?
      path = image.path(:original).sub(%r(\A.*paperclip/), "paperclip/")
      return "#{ENV['ANNICT_FILE_STORAGE_URL']}/#{path}"
    end

    annict_image_url(profile, field, options)
  end
end
