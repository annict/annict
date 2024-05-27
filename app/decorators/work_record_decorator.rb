# typed: false
# frozen_string_literal: true

module WorkRecordDecorator
  def detail_link(options = {})
    title = options.delete(:title).presence || self.title
    path = review_path(user.username, self)
    link_to title, path, options
  end
end
