# frozen_string_literal: true

json.works @works do |work|
  json.id work.id
  json.title work.title
  json.title_kana work.title_kana
  json.media work.media
  json.official_site_url work.official_site_url
  json.wikipedia_url work.wikipedia_url
  json.twitter_username work.twitter_hashtag
  json.episodes_count work.episodes_count
  json.watchers_count work.watchers_count

  json.season do
    json.name work.season.yearly_season_ja
    json.slug work.season.slug
  end
end
