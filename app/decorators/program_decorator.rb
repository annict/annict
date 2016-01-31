class ProgramDecorator < ApplicationDecorator
  include ProgramDecoratorCommon

  def db_detail_link(options = {})
    name = options.delete(:name).presence || name
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_work_program_path(work, self)
    else
      h.new_db_work_draft_program_path(work, id: id)
    end

    h.link_to name, path, options
  end
end
