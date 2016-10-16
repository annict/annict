# frozen_string_literal: true

class OrganizationDecorator < ApplicationDecorator
  include OrganizationDecoratorCommon

  def db_detail_link(options = {})
    name = options.delete(:name).presence || self.name
    h.link_to name, h.edit_db_organization_path(self), options
  end

  def name_link
    h.link_to name, h.organization_path(self)
  end

  def twitter_username_link
    url = "https://twitter.com/#{twitter_username}"
    h.link_to "@#{twitter_username}", url, target: "_blank"
  end

  def wikipedia_url_link
    h.link_to "Wikipedia", wikipedia_url, target: "_blank"
  end

  def url_link
    h.link_to URI.parse(url).host.downcase, url, target: "_blank"
  end
end
