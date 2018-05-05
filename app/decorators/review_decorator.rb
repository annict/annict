# frozen_string_literal: true

class WorkRecordDecorator < ApplicationDecorator
  def detail_link(options = {})
    title = options.delete(:title).presence || self.title
    path = h.review_path(user.username, self)
    h.link_to title, path, options
  end
end
