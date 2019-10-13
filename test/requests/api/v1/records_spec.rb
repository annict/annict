# frozen_string_literal: true

describe "Api::V1::Records" do
  before do
    Timecop.freeze(Time.parse("2017-01-29 08:39:04"))
  end

  after do
    Timecop.return
  end

  let(:access_token) { create(:oauth_access_token) }
  let(:user) { create(:user, :with_profile) }
  let(:work) { create(:work, :with_current_season) }
  let(:episode) { create(:episode, work: work) }
  let!(:record) { create(:episode_record, work: work, episode: episode, user: user) }

  describe "GET /v1/records" do
    before do
      get api("/v1/records", access_token: access_token.token)
    end

    context "when added no parameters" do
      it "responses 200" do
        expect(response.status).to eq(200)
      end

      it "gets record info" do
        expected_hash = {
          "id" => record.id,
          "comment" => "おもしろかった",
          "rating" => 3.0,
          "rating_state" => nil,
          "is_modified" => false,
          "likes_count" => 0,
          "comments_count" => 0,
          "created_at" => "2017-01-28T23:39:04.000Z",
          "user" => {
            "id" => user.id,
            "username" => user.username,
            "name" => user.profile.name,
            "description" => "悟空を倒すために生まれました。よろしくお願いします。",
            "url" => "http://example.com",
            "records_count" => 1,
            "followings_count" => 0,
            "followers_count" => 0,
            "wanna_watch_count" => 0,
            "watching_count" => 0,
            "watched_count" => 0,
            "on_hold_count" => 0,
            "stop_watching_count" => 0,
            "created_at" => "2017-01-28T23:39:04.000Z"
          },
          "work" => {
            "id" => work.id,
            "title" => work.title,
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
          },
          "episode" => {
            "id" => episode.id,
            "number" => episode.raw_number,
            "number_text" => episode.number,
            "sort_number" => episode.sort_number,
            "title" => episode.title,
            "records_count" => 1,
            "record_comments_count" => 0
          }
        }
        actual_hash = json["records"][0]
        actual_hash["user"].delete("avatar_url")
        actual_hash["user"].delete("background_image_url")

        expect(actual_hash).to include(expected_hash)
        expect(expected_hash).to include(actual_hash)
        expect(json["total_count"]).to eq(1)
        expect(json["next_page"]).to eq(nil)
        expect(json["prev_page"]).to eq(nil)
      end
    end
  end
end
