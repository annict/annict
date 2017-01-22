# frozen_string_literal: true

class TipDecorator < ApplicationDecorator
  def local_title
    return title if I18n.locale == :ja
    return title_en if title_en.present?
    title
  end

  def local_body
    return body if I18n.locale == :ja
    return body_en if body_en.present?
    body
  end
end
