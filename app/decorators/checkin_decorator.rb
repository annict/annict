# frozen_string_literal: true

class CheckinDecorator < ApplicationDecorator
  def rating_label
    return if rating.blank?

    h.content_tag :div, class: "rating-label" do
      tags = []
      tags << h.content_tag(:i, nil, class: "fa fa-star")
      tags << h.content_tag(:i, nil, class: "fa #{(rating <= 1) ? 'fa-star-o' : (1 < rating && rating < 2) ? 'fa-star-half-o' : (2 <= rating) ? 'fa-star' : ''}")
      tags << h.content_tag(:i, nil, class: "fa #{(rating <= 2) ? 'fa-star-o' : (2 < rating && rating < 3) ? 'fa-star-half-o' : (3 <= rating) ? 'fa-star' : ''}")
      tags << h.content_tag(:i, nil, class: "fa #{(rating <= 3) ? 'fa-star-o' : (3 < rating && rating < 4) ? 'fa-star-half-o' : (4 <= rating) ? 'fa-star' : ''}")
      tags << h.content_tag(:i, nil, class: "fa #{(rating <= 4) ? 'fa-star-o' : (4 < rating && rating < 5) ? 'fa-star-half-o' : (5 <= rating) ? 'fa-star' : ''}")
      tags << h.content_tag(:span, rating.round(1), class: "text")
      tags.join("").html_safe
    end
  end
end
