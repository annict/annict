# typed: false
# frozen_string_literal: true

module TipDecorator
  def local_title
    return title if I18n.locale == :ja
    return title_en if title_en.present?
    title
  end
end
