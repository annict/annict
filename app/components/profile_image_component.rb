# frozen_string_literal: true

class ProfileImageComponent < ApplicationComponent
  def initialize(view_context, image_url_1x:, alt:, bg_color: "#F1F1F1", class_name: "", img_options: {}, lazy_load: true)
    super view_context
    @image_url_1x = image_url_1x
    @alt = alt
    @bg_color = bg_color
    @class_name = class_name
    @img_options = img_options
    @lazy_load = lazy_load
  end

  def render
    if @lazy_load
      image_tag(@dummy_src, {
        alt: @alt,
        class: img_class_name,
        "data-src": @image_url_1x,
        style: "background-color: #{@bg_color};"
      }.merge(@img_options))
    else
      image_tag(@image_url_1x, {
        alt: @alt,
        class: img_class_name
      }.merge(@img_options))
    end
  end

  private

  def img_class_name
    classes = %w[c-profile-image img-fluid img-thumbnail js-lazy rounded-circle]
    classes << "js-lazy" if @lazy_load
    classes += @class_name.split(" ")
    classes.join(" ")
  end

  def dummy_src
    "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 1 1'%3E%3C/svg%3E"
  end
end
