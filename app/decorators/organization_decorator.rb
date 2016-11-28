# frozen_string_literal: true

class OrganizationDecorator < ApplicationDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || self.name
    h.link_to name, h.edit_db_organization_path(self), options
  end

  def local_name
    return name if I18n.locale == :ja
    return name_en if name_en.present?
    name
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

  def local_name
    return name if I18n.locale == :ja
    return name_en if name_en.present?
    name
  end

  def to_values
    model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :url
        url = send(:url)
        h.link_to(url, url, target: "_blank") if url.present?
      when :wikipedia_url
        wikipedia_url = send(field)
        if wikipedia_url.present?
          h.link_to(URI.decode(wikipedia_url), wikipedia_url, target: "_blank")
        end
      when :twitter_username
        username = send(:twitter_username)
        if username.present?
          url = "https://twitter.com/#{username}"
          h.link_to("@#{username}", url, target: "_blank")
        end
      else
        send(field)
      end

      hash
    end
  end
end
