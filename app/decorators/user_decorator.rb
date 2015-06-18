class UserDecorator < ApplicationDecorator
  def name_link(options = {})
    h.link_to(profile.name, h.user_path(username), options)
  end
end
