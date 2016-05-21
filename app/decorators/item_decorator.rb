# frozen_string_literal: true

class ItemDecorator < ApplicationDecorator
  include ImageMethods
  include ItemDecoratorCommon

  def db_detail_link(options = {})
    name = options.delete(:name).presence || self.name
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_work_item_path(work)
    else
      h.new_db_work_draft_item_path(work, item_id: id)
    end

    h.link_to name, path, options
  end
end
