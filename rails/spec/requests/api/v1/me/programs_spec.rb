# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/me/programs", type: :request do
  it "パラメータを指定しない場合、200ステータスを返すこと" do
    access_token = create(:oauth_access_token)
    user = access_token.owner
    work = create(:work, :with_current_season, watchers_count: 1)
    episode = create(:episode, work:)
    channel = Channel.first
    status = create(:status, kind: "watching", work:, user:)
    slot = create(:slot, work:, episode:, channel:)
    create(:library_entry, user:, work:, status:, program: slot.program)

    data = {
      access_token: access_token.token
    }
    get api("/v1/me/programs", data)

    expect(response.status).to eq(200)
  end

  it "ユーザーが視聴中の作品のスロット情報を取得できること" do
    access_token = create(:oauth_access_token)
    user = access_token.owner
    work = create(:work, :with_current_season, watchers_count: 1)
    episode = create(:episode, work:)
    channel = Channel.first
    status = create(:status, kind: "watching", work:, user:)
    slot = create(:slot, work:, episode:, channel:)
    create(:library_entry, user:, work:, status:, program: slot.program)

    data = {
      access_token: access_token.token
    }
    get api("/v1/me/programs", data)

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

  it "アクセストークンが無効な場合、401ステータスを返すこと" do
    access_token = create(:oauth_access_token)
    user = access_token.owner
    work = create(:work, :with_current_season, watchers_count: 1)
    episode = create(:episode, work:)
    channel = Channel.first
    status = create(:status, kind: "watching", work:, user:)
    slot = create(:slot, work:, episode:, channel:)
    create(:library_entry, user:, work:, status:, program: slot.program)

    data = {
      access_token: "invalid_token"
    }
    get api("/v1/me/programs", data)

    expect(response.status).to eq(401)
  end

  it "視聴中の作品が存在しない場合、空の配列を返すこと" do
    access_token = create(:oauth_access_token)

    data = {
      access_token: access_token.token
    }
    get api("/v1/me/programs", data)

    expect(response.status).to eq(200)
    expect(json["programs"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end

  it "視聴中以外のステータスの作品は表示されないこと" do
    access_token = create(:oauth_access_token)
    user = access_token.owner
    work = create(:work, :with_current_season, watchers_count: 1)
    episode = create(:episode, work:)
    channel = Channel.first
    status = create(:status, kind: "watched", work:, user:)
    slot = create(:slot, work:, episode:, channel:)
    create(:library_entry, user:, work:, status:, program: slot.program)

    data = {
      access_token: access_token.token
    }
    get api("/v1/me/programs", data)

    expect(response.status).to eq(200)
    expect(json["programs"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end
end
