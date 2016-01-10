class PersonDecorator < ApplicationDecorator
  include PersonDecoratorCommon

  def db_detail_link(options = {})
    name = options.delete(:name).presence || self.name
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_person_path(self)
    else
      h.new_db_draft_person_path(person_id: id)
    end

    h.link_to name, path, options
  end
end
