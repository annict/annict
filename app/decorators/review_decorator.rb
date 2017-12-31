# frozen_string_literal: true

class ReviewDecorator < ApplicationDecorator
  def detail_link(options = {})
    title = options.delete(:title).presence || self.title
    path = h.review_path(user.username, self)
    h.link_to title, path, options
  end

  # Do not use helper methods via Draper when the method is used in ActiveJob
  # https://github.com/drapergem/draper/issues/655
  def detail_url
    "#{user.annict_url}/@#{user.username}/reviews/#{id}"
  end

  def tweet_body
    user = review.user
    work_title = work.decorate.local_title
    share_url = detail_url
    share_hashtag = review.work.hashtag_with_hash

    base_body = if user.locale == "ja"
      "%sのレビューを書きました #{share_url} #{share_hashtag}"
    else
      "Reviewed: %s #{share_url} #{share_hashtag}"
    end
    body = base_body % work_title

    return body if body.length <= 140

    work_title = work_title.truncate(work_title.length - (body.length - 140))

    base_body % work_title
  end
end
