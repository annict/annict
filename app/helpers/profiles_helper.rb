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
