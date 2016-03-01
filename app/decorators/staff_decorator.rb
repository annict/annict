# frozen_string_literal: true

class StaffDecorator < ApplicationDecorator
  include StaffDecoratorCommon
  include PersonableDecorator

  def db_detail_link(options = {})
    name = options.delete(:name).presence || name
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_work_staff_path(person, self)
    else
      h.new_db_work_draft_staff_path(person, staff_id: id)
    end

    h.link_to name, path, options
  end
end
