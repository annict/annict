# frozen_string_literal: true

module UserDecorator
  def name_link(options = {})
    link_to(profile.name, annict_url(:profile_url, username), options)
  end

  def role_badge
    return "" unless committer?

    content_tag(:span, class: "badge bg-secondary rounded-pill") do
      role_text
    end
  end

  def name_with_username
    "#{profile.name} (@#{username})"
  end
end
