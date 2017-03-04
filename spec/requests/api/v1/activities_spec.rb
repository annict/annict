# frozen_string_literal: true

describe "Api::V1::Activities" do
  before do
    Timecop.freeze(Time.parse("2017-01-29 08:39:04"))
  end

  after do
    Timecop.return
  end

  describe "GET /v1/activities" do
    let(:user) { create(:user, :with_profile) }
    let(:access_token) { create(:oauth_access_token, owner: user) }
    let!(:record) { create(:checkin, user: user) }
    let!(:activity) do
      create(:activity, user: user, recipient: record.episode, trackable: record)
    end

    before do
      params = {
        access_token: access_token.token
      }

      get api("/v1/activities", params)
    end

    context "when no parameters" do
      it "responses 200" do
        expect(response.status).to eq(200)
      end

      it "gets activity info" do
        expected_hash = {
          "id" => activity.id,
          "user" => {
            "id" => user.id,
            "username" => user.username,
            "name" => user.profile.name,
            "description" => "悟空を倒すために生まれました。よろしくお願いします。",
            "url" => "http://example.com",
            "records_count" => 1,
            "created_at" => "2017-01-28T23:39:04.000Z"
          },
          "action" => activity.action.to_s,
          "created_at" => "2017-01-28T23:39:04.000Z",
          "work" => {
            "id" => record.work.id,
            "title" => record.work.title,
            "title_kana" => record.work.title_kana,
            "media" => "tv",
            "media_text" => "TV",
            "released_on" => "2012-04-05",
            "released_on_about" => "2012年",
            "official_site_url" => "http://example.com",
            "wikipedia_url" => "http://wikipedia.org",
            "twitter_username" => "precure_official",
            "twitter_hashtag" => "precure",
            "episodes_count" => 1,
            "watchers_count" => 0
          },
          "episode" => {
            "id" => record.episode.id,
            "number" => record.episode.raw_number,
            "number_text" => record.episode.number,
            "sort_number" => record.episode.sort_number,
            "title" => record.episode.title,
            "records_count" => 1,
            "record_comments_count" => 0
          },
          "record" => {
            "id" => record.id,
            "comment" => "おもしろかった",
            "rating" => 3.0,
            "is_modified" => false,
            "likes_count" => 0,
            "comments_count" => 0,
            "created_at" => "2017-01-28T23:39:04.000Z"
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
