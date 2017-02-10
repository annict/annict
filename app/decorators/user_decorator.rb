# frozen_string_literal: true

class UserDecorator < ApplicationDecorator
  def name_link(options = {})
    h.link_to(profile.name, h.annict_url(:user_url, username), options)
  end

  def role_badge
    return "" unless committer?

    h.content_tag(:span, class: "u-badge-outline u-badge-outline-default") do
      role_text
    end
  end
end
