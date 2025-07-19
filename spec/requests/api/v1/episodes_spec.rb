# typed: false
# frozen_string_literal: true

describe "Api::V1::Episodes" do
  describe "GET /v1/episodes" do
    it "パラメータなしでアクセスしたとき、200ステータスを返すこと" do
      access_token = create(:oauth_access_token)
      create(:work, :with_current_season, :with_episode)

      get api("/v1/episodes", access_token: access_token.token)

      expect(response.status).to eq(200)
    end

    it "パラメータなしでアクセスしたとき、エピソード情報を取得できること" do
      access_token = create(:oauth_access_token)
      work = create(:work, :with_current_season, :with_episode)

      get api("/v1/episodes", access_token: access_token.token)

      episode = work.episodes.first
      expected_hash = {
        "id" => episode.id,
        "number" => episode.raw_number,
        "number_text" => episode.number,
        "sort_number" => episode.sort_number,
        "title" => episode.title,
        "records_count" => 0,
        "prev_episode" => nil,
        "record_comments_count" => 0,
        "next_episode" => nil,
        "work" => {
          "id" => work.id,
          "title" => work.title,
          "title_en" => "",
          "title_kana" => work.title_kana,
          "media" => "tv",
          "media_text" => "TV",
          "season_name" => "2017-winter",
          "season_name_text" => "2017年冬",
          "released_on" => "2012-04-05",
          "released_on_about" => "2012年",
          "official_site_url" => "http://example.com",
          "wikipedia_url" => "http://wikipedia.org",
          "twitter_username" => "precure_official",
          "twitter_hashtag" => "precure",
          "syobocal_tid" => "",
          "mal_anime_id" => "12345",
          "images" => {
            "recommended_url" => "",
            "facebook" => {
              "og_image_url" => ""
            },
            "twitter" => {
              "mini_avatar_url" => "https://twitter.com/precure_official/profile_image?size=mini",
              "normal_avatar_url" => "https://twitter.com/precure_official/profile_image?size=normal",
              "bigger_avatar_url" => "https://twitter.com/precure_official/profile_image?size=bigger",
              "original_avatar_url" => "https://twitter.com/precure_official/profile_image?size=original",
              "image_url" => ""
            }
          },
          "episodes_count" => 1,
          "watchers_count" => 0,
          "reviews_count" => 0,
          "no_episodes" => false
        }
      }
      expect(json["episodes"][0]).to include(expected_hash)
      expect(json["total_count"]).to eq(1)
      expect(json["next_page"]).to eq(nil)
      expect(json["prev_page"]).to eq(nil)
    end
  end
end
