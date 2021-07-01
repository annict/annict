# frozen_string_literal: true

describe "Api::V1::Me::FollowingActivities" do
  before do
    Timecop.freeze(Time.parse("2017-01-29 08:39:04"))
  end

  after do
    Timecop.return
  end

  describe "GET /v1/me/following_activities" do
    let(:user1) { create(:user, :with_profile) }
    let(:user2) { create(:user, :with_profile) }
    let(:access_token) { create(:oauth_access_token, owner: user1) }
    let!(:follow) { create(:follow, user: user1, following: user2) }
    let!(:anime) { create(:anime) }
    let!(:episode) { create(:episode, anime: anime) }
    let!(:record) { create(:record, user: user2) }
    let!(:episode_record) { create(:episode_record, record: record, user: user2, anime: anime, episode: episode) }
    let!(:activity) { create(:activity, user: user2, itemable: episode_record) }

    before do
      params = {
        access_token: access_token.token
      }

      get api("/v1/me/following_activities", params)
    end

    context "when no parameters" do
      it "responses 200" do
        expect(response.status).to eq(200)
      end

      it "gets activity info" do
        expected_hash = {
          "id" => activity.id,
          "user" => {
            "id" => user2.id,
            "username" => user2.username,
            "name" => user2.profile.name,
            "description" => "悟空を倒すために生まれました。よろしくお願いします。",
            "url" => "http://example.com",
            "records_count" => 1,
            "followings_count" => 0,
            "followers_count" => 1,
            "wanna_watch_count" => 0,
            "watching_count" => 0,
            "watched_count" => 0,
            "on_hold_count" => 0,
            "stop_watching_count" => 0,
            "created_at" => "2017-01-28T23:39:04Z"
          },
          "action" => "create_record",
          "created_at" => "2017-01-28T23:39:04Z",
          "work" => {
            "id" => anime.id,
            "title" => anime.title,
            "title_en" => "",
            "title_kana" => anime.title_kana,
            "media" => "tv",
            "media_text" => "TV",
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
            "record_comments_count" => 1
          },
          "record" => {
            "id" => episode_record.id,
            "comment" => "おもしろかった",
            "rating" => 3.0,
            "rating_state" => nil,
            "is_modified" => false,
            "likes_count" => 0,
            "comments_count" => 0,
            "created_at" => "2017-01-28T23:39:04Z"
          }
        }
        actual_hash = json["activities"][0].stringify_keys
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
