class WorkOrganizationDecorator < ApplicationDecorator
  include WorkOrganizationDecoratorCommon

  def db_detail_link(options = {})
    name = options.delete(:name).presence || name
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_work_work_organization_path(work, self)
    else
      h.new_db_work_draft_work_organization_path(work,
        work_organization_id: id)
    end

    h.link_to name, path, options
  end
end
