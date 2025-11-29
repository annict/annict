# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/reviews", type: :request do
  it "パラメーターを指定しない場合、200ステータスを返すこと" do
    Timecop.freeze(Time.parse("2017-01-29 08:39:04"))

    access_token = create(:oauth_access_token)
    user = create(:user, :with_profile)
    work = create(:work, :with_current_season)
    record = create(:record, user:, work:)
    create(:work_record, record:, work:, user:)

    get api("/v1/reviews", access_token: access_token.token)

    expect(response.status).to eq(200)

    Timecop.return
  end

  it "パラメーターを指定しない場合、レビュー情報を取得できること" do
    Timecop.freeze(Time.parse("2017-01-29 08:39:04"))

    access_token = create(:oauth_access_token)
    user = create(:user, :with_profile)
    work = create(:work, :with_current_season)
    record = create(:record, user:, work:)
    work_record = create(:work_record, record:, work:, user:)

    get api("/v1/reviews", access_token: access_token.token)

    expected_hash = {
      "id" => work_record.id,
      "title" => "",
      "body" => "おもしろかった",
      "rating_animation_state" => nil,
      "rating_music_state" => nil,
      "rating_story_state" => nil,
      "rating_character_state" => nil,
      "rating_overall_state" => nil,
      "likes_count" => 0,
      "impressions_count" => 0,
      "modified_at" => nil,
      "created_at" => "2017-01-28T23:39:04Z",
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
        "created_at" => "2017-01-28T23:39:04Z"
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
        "episodes_count" => 0,
        "watchers_count" => 0,
        "reviews_count" => 1,
        "no_episodes" => false
      }
    }
    actual_hash = json["reviews"][0]
    actual_hash["user"].delete("avatar_url")
    actual_hash["user"].delete("background_image_url")

    expect(actual_hash).to include(expected_hash)
    expect(expected_hash).to include(actual_hash)
    expect(json["total_count"]).to eq(1)
    expect(json["next_page"]).to eq(nil)
    expect(json["prev_page"]).to eq(nil)

    Timecop.return
  end

  it "アクセストークンが無効な場合、401ステータスを返すこと" do
    get api("/v1/reviews", access_token: "invalid_token")

    expect(response.status).to eq(401)
  end

  it "レビューが存在しない場合、空の配列を返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/reviews", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["reviews"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end
end
