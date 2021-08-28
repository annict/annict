# frozen_string_literal: true

describe "GET /v1/me/programs" do
  let(:access_token) { create(:oauth_access_token) }
  let(:user) { access_token.owner }

  context "パラメータを指定しないとき" do
    let(:work) { create(:work, :with_current_season, watchers_count: 1) }
    let(:episode) { create(:episode, work: work) }
    let(:channel) { Channel.first }
    let(:status) { create(:status, kind: "watching", work: work, user: user) }
    let!(:slot) { create(:slot, work: work, episode: episode, channel: channel) }
    let!(:library_entry) { create(:library_entry, user: user, work: work, status: status, program: slot.program) }

    it "レスポンスが返ること" do
      get api("/v1/me/programs", {
        access_token: access_token.token
      })

      expect(response.status).to eq(200)

      work = episode.work
      expected_hash = {
        "id" => slot.id,
        "started_at" => "2017-01-28T15:00:00.000Z",
        "is_rebroadcast" => false,
        "channel" => {
          "id" => channel.id,
          "name" => channel.name
        },
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

  context "filter_unwatched=true を渡したとき" do
    let!(:work) { create(:work) }
    let!(:channel) { Channel.first }
    let!(:program) { create(:program, work: work, channel: channel) }
    let!(:episode_1) { create(:episode, work: work) }
    let!(:episode_1_slot) { create(:slot, channel: channel, program: program, work: work, episode: episode_1) }
    let!(:episode_2) { create(:episode, work: work) }
    let!(:episode_2_slot) { create(:slot, channel: channel, program: program, work: work, episode: episode_2) }
    let!(:status) { create(:status, kind: "watching", work: work, user: user) }
    let!(:library_entry) { create(:library_entry, user: user, work: work, status: status, program: program, watched_episode_ids: [episode_1.id]) }

    it "未視聴のエピソードに紐付く放送予定だけが返ること" do
      get api("/v1/me/programs", {
        access_token: access_token.token,
        filter_unwatched: true
      })

      expect(response.status).to eq(200)
      expect(json["programs"].size).to eq 1
      expect(json["programs"][0]["id"]).to eq episode_2_slot.id
      expect(json["total_count"]).to eq(1)
    end
  end
end
