# frozen_string_literal: true

module UrlHelper
  def link_with_domain(url)
    link_to Addressable::URI.parse(url).host.downcase, url, target: "_blank", rel: "noopener"
  end

  def twitter_profile_link(username)
    link_to "@#{username}", "https://twitter.com/#{username}", target: "_blank", rel: "noopener"
  end

  def twitter_hashtag_link(hashtag)
    link_to "##{hashtag}", "https://twitter.com/search?q=%23#{hashtag}", target: "_blank", rel: "noopener"
  end

  def syobocal_link(syobocal_tid)
    link_to syobocal_tid, "http://cal.syoboi.jp/tid/#{syobocal_tid}", target: "_blank", rel: "noopener"
  end

  def mal_link(mal_anime_id)
    link_to mal_anime_id, "https://myanimelist.net/anime#{mal_anime_id}", target: "_blank", rel: "noopener"
  end

  def annict_url(method, *args, **options)
    options = options.merge(subdomain: nil)
    options = options.merge(protocol: "https") if Rails.env.production?
    send(method, *args, options)
  end
end
