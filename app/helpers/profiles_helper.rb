module ProfilesHelper
  def profile_background_image_url(model)
    size = browser.mobile? ? '640x650e' : '500x325e'
    accessor_name = model.background_image.present? ? :background_image : :avatar

    thumb_url(model, accessor_name, size)
  end
end
