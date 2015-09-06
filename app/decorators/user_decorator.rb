class UserDecorator < ApplicationDecorator
  def name_link(options = {})
    h.link_to(profile.name, h.user_path(username), options)
  end

  def role_label
    if role.editor? || role.admin?
      h.content_tag(:span, class: "role-label label label-default") do
        role_text
      end
    end
  end
end
