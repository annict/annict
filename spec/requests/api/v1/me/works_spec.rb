# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/me/works", type: :request do
  it "200レスポンスを返すこと" do
    access_token = create(:oauth_access_token)
    user = access_token.owner
    work = create(:work, :with_current_season, watchers_count: 1)
    status = create(:status, kind: "watching", work: work, user: user)
    create(:library_entry, user: user, work: work, status: status)

    data = {
      access_token: access_token.token
    }
    get api("/v1/me/works", data)

    expect(response.status).to eq(200)
  end

  it "ユーザーがステータスを更新した作品を取得できること" do
    access_token = create(:oauth_access_token)
    user = access_token.owner
    work = create(:work, :with_current_season, watchers_count: 1)
    status = create(:status, kind: "watching", work: work, user: user)
    create(:library_entry, user: user, work: work, status: status)

    data = {
      access_token: access_token.token
    }
    get api("/v1/me/works", data)

    expected_hash = {
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
      "episodes_count" => 0,
      "watchers_count" => 1,
      "status" => {
        "kind" => "watching"
      },
      "reviews_count" => 0,
      "no_episodes" => false
    }
    expect(json["works"][0].stringify_keys).to include(expected_hash)
    expect(expected_hash).to include(json["works"][0].stringify_keys)
    expect(json["total_count"]).to eq(1)
    expect(json["next_page"]).to eq(nil)
    expect(json["prev_page"]).to eq(nil)
  end

  it "アクセストークンが不正な場合、401を返すこと" do
    get api("/v1/me/works", {access_token: "invalid_token"})

    expect(response.status).to eq(401)
  end

  it "アクセストークンがない場合、401を返すこと" do
    get api("/v1/me/works", {})

    expect(response.status).to eq(401)
  end
end
