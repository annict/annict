# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/users", type: :request do
  it "パラメーターを指定しない場合、200ステータスを返すこと" do
    Timecop.freeze(Time.parse("2017-01-29 08:39:04"))

    user = create(:user, :with_profile)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, owner: user, application:)

    get api("/v1/users", access_token: access_token.token)

    expect(response.status).to eq(200)

    Timecop.return
  end

  it "パラメーターを指定しない場合、ユーザー情報を取得できること" do
    Timecop.freeze(Time.parse("2017-01-29 08:39:04"))

    user = create(:user, :with_profile)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, owner: user, application:)

    get api("/v1/users", access_token: access_token.token)

    expected_hash = {
      "id" => user.id,
      "username" => user.username,
      "name" => user.profile.name,
      "description" => "悟空を倒すために生まれました。よろしくお願いします。",
      "url" => "http://example.com",
      "records_count" => 0,
      "followings_count" => 0,
      "followers_count" => 0,
      "wanna_watch_count" => 0,
      "watching_count" => 0,
      "watched_count" => 0,
      "on_hold_count" => 0,
      "stop_watching_count" => 0,
      "created_at" => "2017-01-28T23:39:04Z"
    }
    actual_hash = json["users"][0].stringify_keys.except(
      "avatar_url",
      "background_image_url"
    )
    expect(actual_hash).to include(expected_hash)
    expect(expected_hash).to include(actual_hash)
    expect(json["total_count"]).to eq(1)
    expect(json["next_page"]).to eq(nil)
    expect(json["prev_page"]).to eq(nil)

    Timecop.return
  end

  it "アクセストークンが無効な場合、401ステータスを返すこと" do
    get api("/v1/users", access_token: "invalid_token")

    expect(response.status).to eq(401)
  end
end
