# frozen_string_literal: true

module UrlHelper
  def link_with_domain(url)
    link_to URI.parse(URI.encode(url)).host.downcase, url, target: "_blank", rel: "noopener"
  end

  def annict_url(method, *args, **options)
    options = options.merge(subdomain: nil)
    options = options.merge(protocol: "https") if Rails.env.production?
    send(method, *args, options)
  end
end
