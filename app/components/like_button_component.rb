# frozen_string_literal: true

class LikeButtonComponent < ApplicationComponent
  def initialize(resource_name:, resource_id:, likes_count:, page_category:, class_name: "", init_is_liked: false)
    @resource_name = resource_name
    @resource_id = resource_id
    @likes_count = likes_count
    @page_category = page_category
    @class_name = class_name
    @init_is_liked = init_is_liked == true
  end

  private

  def like_button_class_name
    classes = %w[d-inline-block u-fake-link]
    classes += @class_name.split(" ")
    classes.uniq.join(" ")
  end
end
