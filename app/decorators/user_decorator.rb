class UserDecorator < ApplicationDecorator
  def name_link(options = {})
    h.link_to(profile.name, h.user_path(username), options)
  end

  def role_label
    return "" unless committer?

    h.content_tag(:span, class: "label label-default c-label--transparent") do
      role_text
    end
  end
end
