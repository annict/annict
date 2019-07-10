# frozen_string_literal: true

module WorkStructDecorator
  def local_synopsis(raw: false)
    text = case I18n.locale
    when :en then synopsis_en
    else synopsis
    end

    return if text.blank?

    raw ? text : simple_format(text)
  end
end
