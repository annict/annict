# frozen_string_literal: true

describe "Api::V1::Me::Programs" do
  let(:access_token) { create(:oauth_access_token) }
  let(:work) { create(:work, :with_current_season) }
  let(:episode) { create(:episode, work: work) }
  let!(:status) do
    create(:status, kind: "watching", work: work, user: access_token.owner)
  end
  let(:channel_work) do
    create(:channel_work, user: access_token.owner, work: work)
  end
  let!(:program) { create(:program, work: work, episode: episode, channel: channel_work.channel) }

  describe "GET /v1/me/programs" do
    before do
      data = {
        access_token: access_token.token
      }
      get api("/v1/me/programs", data)
    end

    it "responses 200" do
      expect(response.status).to eq(200)
    end

    it "gets programs which user is watching" do
      channel = channel_work.channel
      work = episode.work
      expected_hash = {
        "id" => program.id,
        "started_at" => "2017-01-28T15:00:00.000Z",
        "is_rebroadcast" => false,
        "channel" => {
          "id" => channel.id,
          "name" => channel.name
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
          "watchers_count" => 1,
          "reviews_count" => 0,
          "no_episodes" => false
        },
        "episode" => {
          "id" => episode.id,
          "number" => episode.raw_number,
          "number_text" => episode.number,
          "sort_number" => episode.sort_number,
          "title" => episode.title,
          "records_count" => 0,
          "record_comments_count" => 0
        }
      }
      expect(json["programs"][0]).to include(expected_hash)
      expect(json["total_count"]).to eq(1)
      expect(json["next_page"]).to eq(nil)
      expect(json["prev_page"]).to eq(nil)
    end
  end
end
