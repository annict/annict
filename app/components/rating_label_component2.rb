# frozen_string_literal: true

class RatingLabelComponent2 < ApplicationComponent2
  def initialize(view_context, rating:, advanced_rating: nil, class_name: "")
    super view_context
    @rating = rating.downcase.to_sym
    @advanced_rating = advanced_rating
    @class_name = class_name
  end

  def render
    build_html do |h|
      h.tag :span, class: label_class_name do
        h.tag :i, class: icon_class_name
        h.text label_text
      end
    end
  end

  private

  def label_class_name
    classes = %w[badge]
    classes += @class_name.split(" ")
    classes << badge_class_name
    classes.join(" ")
  end

  def icon_class_name
    classes = %w[far]
    classes << {
      great: "fa-heart",
      good: "fa-thumbs-up",
      average: "fa-meh",
      bad: "fa-thumbs-down"
    }[@rating]
    classes.join(" ")
  end

  def badge_class_name
    "u-badge-#{@rating}"
  end

  def label_text
    text = t "enumerize.episode_record.rating_state.#{@rating}"

    if @advanced_rating
      return "#{text} (#{@advanced_rating})"
    end

    text
  end
end
