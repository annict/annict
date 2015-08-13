class WorkDecorator < ApplicationDecorator
  include WorkDecoratorCommon

  def twitter_username_link
    url = "https://twitter.com/#{twitter_username}"
    h.link_to "@#{twitter_username}", url, target: "_blank"
  end

  def twitter_hashtag_link
    url = URI.encode("https://twitter.com/search?q=##{twitter_hashtag}&src=hash")
    h.link_to "##{twitter_hashtag}", url, target: "_blank"
  end

  def syobocal_link(title = nil)
    title = title.presence || sc_tid
    h.link_to title, syobocal_url, target: "_blank"
  end

  def item_image_url(size)
    if main_item.present? && main_item.id != 1
      h.tombo_thumb_url(main_item, :tombo_image, size)
    else
      path = h.image_path('no-image.jpg')
      "#{ENV['ANNICT_TOMBO_URL']}/#{size}#{path}"
    end
  end
end
