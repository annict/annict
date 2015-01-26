module ProfilesHelper
  def profile_background_image_url(profile, mini: false)
    size = if mini
      browser.mobile? ? '640x320e' : '500x250e'
    else
      browser.mobile? ? '640x650e' : '500x325e'
    end
    accessor_name = profile.background_image.present? ? :background_image : :avatar

    thumb_url(profile, accessor_name, size)
  end
end
