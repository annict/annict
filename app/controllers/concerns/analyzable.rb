# frozen_string_literal: true

module Analyzable
  extend ActiveSupport::Concern

  def ga_client
    @ga_client ||= Annict::Analytics::Client.new(request, current_user)
  end

  def ga_tracking_id(request)
    return ENV.fetch("GA_TRACKING_ID_API") if request.subdomain == "api"

    case [request.subdomain, request.domain].select(&:present?).join(".")
    when ENV.fetch("ANNICT_JP_DOMAIN")
      ENV.fetch("GA_TRACKING_ID_JP")
    else
      ENV.fetch("GA_TRACKING_ID")
    end
  end
end
