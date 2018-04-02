# frozen_string_literal: true

module LocalHelper
  def local_domain(locale: I18n.locale)
    case locale.to_s
      when "ja"
        ENV.fetch("ANNICT_JP_DOMAIN")
      else
        ENV.fetch("ANNICT_DOMAIN")
    end
  end
end
