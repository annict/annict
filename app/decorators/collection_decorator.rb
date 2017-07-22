# frozen_string_literal: true

class CollectionDecorator < ApplicationDecorator
  def image_url(options = {})
    return "/no-image.jpg" if collection_items.blank?

    h.ann_image_url collection_items.first.work.work_image, :attachment, options
  end
end
