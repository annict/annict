# typed: false
# frozen_string_literal: true

module GaHelper
  def ga_tracking_id(request)
    return ENV.fetch("GA_TRACKING_ID_API") if request.subdomain == "api"

    case [request.subdomain, request.domain].select(&:present?).join(".")
    when ENV.fetch("ANNICT_DOMAIN")
      ENV.fetch("GA_TRACKING_ID_JP")
    else
      ENV.fetch("GA_TRACKING_ID")
    end
  end
end
