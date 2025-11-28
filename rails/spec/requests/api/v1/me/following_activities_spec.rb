# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/me/following_activities", type: :request do
  before do
    Timecop.freeze(Time.parse("2017-01-29 08:39:04"))
  end

  after do
    Timecop.return
  end

  it "パラメータを指定しない場合、200ステータスを返すこと" do
    user1 = create(:user, :with_profile)
    user2 = create(:user, :with_profile)
    access_token = create(:oauth_access_token, owner: user1)
    create(:follow, user: user1, following: user2)
    work = create(:work)
    episode = create(:episode, work:)
    record = create(:record, user: user2)
    episode_record = create(:episode_record, record:, user: user2, work:, episode:)
    create(:activity, user: user2, itemable: episode_record)

    params = {
      access_token: access_token.token
    }

    get api("/v1/me/following_activities", params)

    expect(response.status).to eq(200)
  end

  it "パラメータを指定しない場合、アクティビティ情報を取得できること" do
    user1 = create(:user, :with_profile)
    user2 = create(:user, :with_profile)
    access_token = create(:oauth_access_token, owner: user1)
    create(:follow, user: user1, following: user2)
    work = create(:work)
    episode = create(:episode, work:)
    record = create(:record, user: user2)
    episode_record = create(:episode_record, record:, user: user2, work:, episode:)
    activity = create(:activity, user: user2, itemable: episode_record)

    params = {
      access_token: access_token.token
    }

    get api("/v1/me/following_activities", params)

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
        "id" => work.id,
        "title" => work.title,
        "title_en" => "",
        "title_kana" => work.title_kana,
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

  it "アクセストークンが無効な場合、401ステータスを返すこと" do
    user1 = create(:user, :with_profile)
    user2 = create(:user, :with_profile)
    create(:follow, user: user1, following: user2)
    work = create(:work)
    episode = create(:episode, work:)
    record = create(:record, user: user2)
    episode_record = create(:episode_record, record:, user: user2, work:, episode:)
    create(:activity, user: user2, itemable: episode_record)

    params = {
      access_token: "invalid_token"
    }

    get api("/v1/me/following_activities", params)

    expect(response.status).to eq(401)
  end

  it "フォローしているユーザーのアクティビティが存在しない場合、空の配列を返すこと" do
    user1 = create(:user, :with_profile)
    user2 = create(:user, :with_profile)
    access_token = create(:oauth_access_token, owner: user1)
    create(:follow, user: user1, following: user2)

    params = {
      access_token: access_token.token
    }

    get api("/v1/me/following_activities", params)

    expect(response.status).to eq(200)
    expect(json["activities"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end

  it "フォローしていないユーザーのアクティビティは表示されないこと" do
    user1 = create(:user, :with_profile)
    user2 = create(:user, :with_profile)
    user3 = create(:user, :with_profile)
    access_token = create(:oauth_access_token, owner: user1)
    create(:follow, user: user1, following: user2)
    work = create(:work)
    episode = create(:episode, work:)

    # user2のアクティビティ（フォローしているので表示される）
    record2 = create(:record, user: user2)
    episode_record2 = create(:episode_record, record: record2, user: user2, work:, episode:)
    create(:activity, user: user2, itemable: episode_record2)

    # user3のアクティビティ（フォローしていないので表示されない）
    record3 = create(:record, user: user3)
    episode_record3 = create(:episode_record, record: record3, user: user3, work:, episode:)
    create(:activity, user: user3, itemable: episode_record3)

    params = {
      access_token: access_token.token
    }

    get api("/v1/me/following_activities", params)

    expect(response.status).to eq(200)
    expect(json["activities"].length).to eq(1)
    expect(json["activities"][0]["user"]["id"]).to eq(user2.id)
    expect(json["total_count"]).to eq(1)
  end
end
