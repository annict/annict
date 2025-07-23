# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/me", type: :request do
  it "200レスポンスを返すこと" do
    Timecop.freeze(Time.parse("2017-01-29 08:39:04"))
    user = create(:user, :with_profile)
    access_token = create(:oauth_access_token, owner: user)

    data = {
      access_token: access_token.token
    }
    get api("/v1/me", data)

    expect(response.status).to eq(200)

    Timecop.return
  end

  it "ユーザー情報を取得できること" do
    Timecop.freeze(Time.parse("2017-01-29 08:39:04"))
    user = create(:user, :with_profile)
    access_token = create(:oauth_access_token, owner: user)

    data = {
      access_token: access_token.token
    }
    get api("/v1/me", data)

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
      "created_at" => "2017-01-28T23:39:04Z",
      "email" => user.email,
      "notifications_count" => 0
    }
    actual_hash = json.stringify_keys.except(
      "avatar_url",
      "background_image_url"
    )

    expect(actual_hash).to include(expected_hash)
    expect(expected_hash).to include(actual_hash)

    Timecop.return
  end

  it "アクセストークンが不正な場合、401を返すこと" do
    get api("/v1/me", {access_token: "invalid_token"})

    expect(response.status).to eq(401)
  end

  it "アクセストークンがない場合、401を返すこと" do
    get api("/v1/me", {})

    expect(response.status).to eq(401)
  end
end
