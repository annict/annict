# frozen_string_literal: true

module LocalHelper
  def local_domain(locale: I18n.locale)
    case locale.to_s
    when "ja"
      ENV.fetch("ANNICT_DOMAIN")
    else
      ENV.fetch("ANNICT_EN_DOMAIN")
    end
  end

  def local_url(locale: I18n.locale)
    return ENV.fetch("ANNICT_URL") if locale.to_s == "ja"

    ENV.fetch("ANNICT_EN_URL")
  end
end
