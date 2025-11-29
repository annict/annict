# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/works", type: :request do
  it "パラメータを指定しない場合、200ステータスを返すこと" do
    access_token = create(:oauth_access_token)
    create(:work, :with_current_season, :with_episode)

    get api("/v1/works", access_token: access_token.token)

    expect(response.status).to eq(200)
  end

  it "パラメータを指定しない場合、作品情報を取得できること" do
    access_token = create(:oauth_access_token)
    work = create(:work, :with_current_season, :with_episode)

    get api("/v1/works", access_token: access_token.token)

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

  it "アクセストークンが無効な場合、401ステータスを返すこと" do
    create(:work, :with_current_season, :with_episode)

    get api("/v1/works", access_token: "invalid_token")

    expect(response.status).to eq(401)
  end

  it "作品が存在しない場合、空の配列を返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/works", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["works"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end
end
