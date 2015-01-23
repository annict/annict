module ProfilesHelper
  def background_image_url(model)
    size = browser.mobile? ? '640x650c' : '500x325c'
    accessor_name = model.background_image.present? ? :background_image : :avatar

    thumb_url(model, accessor_name, size)
  end
end
