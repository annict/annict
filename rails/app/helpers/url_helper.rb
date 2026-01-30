# typed: false
# frozen_string_literal: true

module UrlHelper
  def link_with_domain(url)
    link_to Addressable::URI.parse(url).host.downcase, url, target: "_blank", rel: "noopener"
  end

  def annict_url(method, *, **options)
    options = options.merge(subdomain: nil)
    options = options.merge(protocol: "https") if Rails.env.production?
    send(method, *, options)
  end

  def current_path_with_query
    [request.path, request.query_string].select(&:present?).join("?")
  end

  def ics_calendar_url(username)
    "#{annict_url(:root_url)}@#{username}/ics"
  end

  def ics_calendar_alt_url(username)
    "#{annict_url(:root_url)}ics?username=#{username}"
  end
end
