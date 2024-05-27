# typed: false
# frozen_string_literal: true

describe "Api::V1::Works" do
  let(:access_token) { create(:oauth_access_token) }
  let!(:work) { create(:work, :with_current_season, :with_episode) }

  describe "GET /v1/works" do
    before do
      get api("/v1/works", access_token: access_token.token)
    end

    context "when added no parameters" do
      it "responses 200" do
        expect(response.status).to eq(200)
      end

      it "gets work info" do
        expected_hash = {
          "id" => work.id,
          "title" => work.title,
          "title_kana" => work.title_kana,
          "title_en" => work.title_en,
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
        expect(json["works"][0].stringify_keys).to include(expected_hash)
        expect(expected_hash).to include(json["works"][0].stringify_keys)
        expect(json["total_count"]).to eq(1)
        expect(json["next_page"]).to eq(nil)
        expect(json["prev_page"]).to eq(nil)
      end
    end
  end
end
