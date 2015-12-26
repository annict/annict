class CastDecorator < ApplicationDecorator
  include CastDecoratorCommon

  def db_detail_link(options = {})
    name = options.delete(:name).presence || name
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_person_cast_path(person, self)
    else
      h.new_db_person_draft_cast_path(person, cast_id: id)
    end

    h.link_to name, path, options
  end
end
