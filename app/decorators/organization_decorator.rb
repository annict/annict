# frozen_string_literal: true

module OrganizationDecorator
  include PersonOrgDecoratorCommon

  def db_detail_link(options = {})
    name = options.delete(:name).presence || self.name
    link_to name, db_edit_organization_path(self), options
  end

  def grid_description(staff)
    staff.decorate.role_name
  end

  def to_values
    self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :url
        url = send(:url)
        link_to(url, url, target: "_blank", rel: "noopener") if url.present?
      when :wikipedia_url
        wikipedia_url = send(field)
        if wikipedia_url.present?
          link_to(wikipedia_url, wikipedia_url, target: "_blank", rel: "noopener")
        end
      when :twitter_username
        username = send(:twitter_username)
        if username.present?
          url = "https://twitter.com/#{username}"
          link_to("@#{username}", url, target: "_blank", rel: "noopener")
        end
      else
        send(field)
      end

      hash
    end
  end
end
