# frozen_string_literal: true

module WorkEntityDecorator
  def media_text
    t "enumerize.work.media.#{media}"
  end

  def twitter_profile_link
    link_to "@#{twitter_username}", "https://twitter.com/#{twitter_username}", target: "_blank", rel: "noopener"
  end

  def twitter_hashtag_link
    link_to "##{twitter_hashtag}", "https://twitter.com/search?q=%23#{twitter_hashtag}", target: "_blank", rel: "noopener"
  end

  def syobocal_link
    link_to syobocal_tid, "http://cal.syoboi.jp/tid/#{syobocal_tid}", target: "_blank", rel: "noopener"
  end

  def mal_link
    link_to mal_anime_id, "https://myanimelist.net/anime/#{mal_anime_id}", target: "_blank", rel: "noopener"
  end
end
