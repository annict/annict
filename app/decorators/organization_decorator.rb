class OrganizationDecorator < ApplicationDecorator
  include OrganizationDecoratorCommon

  def db_detail_link(options = {})
    name = options.delete(:name).presence || self.name
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_organization_path(self)
    else
      h.new_db_draft_organization_path(organization_id: id)
    end

    h.link_to name, path, options
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
