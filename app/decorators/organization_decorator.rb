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
end
