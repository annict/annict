# frozen_string_literal: true

class CastDecorator < ApplicationDecorator
  include CastDecoratorCommon

  def db_detail_link(options = {})
    name = options.delete(:name).presence || name
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_work_cast_path(work, self)
    else
      h.new_db_work_draft_cast_path(work, cast_id: id)
    end

    h.link_to name, path, options
  end

  def name_with_old
    return name if name == person.name
    "#{name} (#{person.name})"
  end

  def name_with_old_link
    h.link_to name_with_old, h.person_path(person)
  end
end
