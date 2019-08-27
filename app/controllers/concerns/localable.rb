# frozen_string_literal: true

module Localable
  private

  def locale_ja?
    locale.to_s == "ja"
  end

  def locale_en?
    locale.to_s == "en"
  end

  def local_url(locale: I18n.locale)
    case locale.to_s
    when "ja"
      ENV.fetch("ANNICT_JP_URL")
    else
      ENV.fetch("ANNICT_URL")
    end
  end
end
