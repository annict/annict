class WorkOrganizationDecorator < ApplicationDecorator
  include WorkOrganizationDecoratorCommon

  def db_detail_link(options = {})
    name = options.delete(:name).presence || name
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_organization_work_organization_path(organization, self)
    else
      h.new_db_organization_draft_work_organization_path(organization,
        work_organization_id: id)
    end

    h.link_to name, path, options
  end
end
