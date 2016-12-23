# frozen_string_literal: true

module UrlHelper
  def link_with_domain(url)
    link_to URI.parse(url).host.downcase, url, target: "_blank"
  end
end
