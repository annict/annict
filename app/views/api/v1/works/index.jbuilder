# frozen_string_literal: true

json.works @works do |work|
  json.id work.id if @params.fields_contain?("id")
  json.title work.title if @params.fields_contain?("title")
  json.title_kana work.title_kana if @params.fields_contain?("title_kana")
  json.media work.media if @params.fields_contain?("media")
  json.media_text work.media_text if @params.fields_contain?("media_text")
  json.season_name work.season.slug if @params.fields_contain?("season_name")
  json.season_name_text work.season.yearly_season_ja if @params.fields_contain?("season_name_text")
  json.released_on work.released_at&.strftime("%Y-%m-%d").presence || "" if @params.fields_contain?("released_on")
  json.released_on_about work.released_at_about.presence || "" if @params.fields_contain?("released_on_about")
  json.official_site_url work.official_site_url.presence || "" if @params.fields_contain?("official_site_url")
  json.wikipedia_url work.wikipedia_url.presence || "" if @params.fields_contain?("wikipedia_url")
  json.twitter_username work.twitter_username.presence || "" if @params.fields_contain?("twitter_username")
  json.twitter_hashtag work.twitter_hashtag.presence || "" if @params.fields_contain?("twitter_hashtag")
  json.episodes_count work.episodes_count if @params.fields_contain?("episodes_count")
  json.watchers_count work.watchers_count if @params.fields_contain?("watchers_count")
end
