# frozen_string_literal: true

class ReviewDecorator < ApplicationDecorator
  def detail_link(options = {})
    title = options.delete(:title).presence || self.title
    path = h.review_path(user.username, self)
    h.link_to title, path, options
  end

  def detail_url
    h.review_url(user.username, self)
  end

  def tweet_body
    work_title = work.title
    share_url = h.review_url(user.username, self)
    share_hashtag = review.work.hashtag_with_hash
    review_title = title

    body = "#{work_title}のレビュー「#{review_title}」を書きました #{share_url} #{share_hashtag}"
    return body if body.length <= 140

    review_title = review_title.truncate(review_title.length - (body.length - 140))

    "#{work_title}のレビュー「#{review_title}」を書きました #{share_url} #{share_hashtag}"
  end
end
