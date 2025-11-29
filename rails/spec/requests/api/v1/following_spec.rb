# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/following", type: :request do
  before do
    Timecop.freeze(Time.parse("2017-01-29 08:39:04"))
  end

  after do
    Timecop.return
  end

  it "filter_usernameパラメータが指定されている場合、ステータスコード200を返すこと" do
    user1 = create(:user, :with_profile)
    user2 = create(:user, :with_profile)
    access_token = create(:oauth_access_token, owner: user1)
    create(:follow, user: user1, following: user2)

    params = {
      access_token: access_token.token,
      filter_username: user1.username
    }

    get api("/v1/following", params)

    expect(response.status).to eq(200)
  end

  it "filter_usernameパラメータが指定されている場合、フォローしているユーザーの情報を取得できること" do
    user1 = create(:user, :with_profile)
    user2 = create(:user, :with_profile)
    access_token = create(:oauth_access_token, owner: user1)
    create(:follow, user: user1, following: user2)

    params = {
      access_token: access_token.token,
      filter_username: user1.username
    }

    get api("/v1/following", params)

    expected_hash = {
      "id" => user2.id,
      "username" => user2.username,
      "name" => user2.profile.name,
      "description" => "悟空を倒すために生まれました。よろしくお願いします。",
      "url" => "http://example.com",
      "records_count" => 0,
      "followings_count" => 0,
      "followers_count" => 1,
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
  end

  it "アクセストークンが提供されていない場合、401エラーを返すこと" do
    get api("/v1/following")
    expect(response.status).to eq(401)
  end

  it "無効なアクセストークンが提供された場合、401エラーを返すこと" do
    get api("/v1/following", access_token: "invalid_token")
    expect(response.status).to eq(401)
  end

  it "フォローしているユーザーが存在しない場合、空の配列を返すこと" do
    user = create(:user)
    access_token = create(:oauth_access_token, owner: user)

    get api("/v1/following", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["users"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end

  it "存在しないusernameでフィルタした場合、空の結果を返すこと" do
    user = create(:user)
    access_token = create(:oauth_access_token, owner: user)

    get api("/v1/following", access_token: access_token.token, filter_username: "nonexistent")

    expect(response.status).to eq(200)
    expect(json["users"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end

  it "filter_user_idパラメータが指定されている場合、正しいユーザーのフォロー情報を取得できること" do
    user1 = create(:user, :with_profile)
    user2 = create(:user, :with_profile)
    user3 = create(:user, :with_profile)
    access_token = create(:oauth_access_token, owner: user1)

    create(:follow, user: user1, following: user2)
    create(:follow, user: user3, following: user2)

    get api("/v1/following", access_token: access_token.token, filter_user_id: user1.id)

    expect(response.status).to eq(200)
    expect(json["users"].size).to eq(1)
    expect(json["users"][0]["id"]).to eq(user2.id)
    expect(json["total_count"]).to eq(1)
  end

  it "複数のフォロー関係が存在する場合、正しく取得できること" do
    user1 = create(:user, :with_profile)
    user2 = create(:user, :with_profile)
    user3 = create(:user, :with_profile)
    access_token = create(:oauth_access_token, owner: user1)

    create(:follow, user: user1, following: user2)
    create(:follow, user: user1, following: user3)

    get api("/v1/following", access_token: access_token.token, filter_user_id: user1.id)

    expect(response.status).to eq(200)
    expect(json["users"].size).to eq(2)
    expect(json["total_count"]).to eq(2)
  end

  it "pageパラメータが0の場合、400エラーを返すこと" do
    user = create(:user)
    access_token = create(:oauth_access_token, owner: user)

    get api("/v1/following", access_token: access_token.token, page: 0)

    expect(response.status).to eq(400)
  end

  it "per_pageパラメータが上限を超える場合、400エラーを返すこと" do
    user = create(:user)
    access_token = create(:oauth_access_token, owner: user)

    get api("/v1/following", access_token: access_token.token, per_page: 51)

    expect(response.status).to eq(400)
  end

  it "sort_idパラメータが無効な値の場合、400エラーを返すこと" do
    user = create(:user)
    access_token = create(:oauth_access_token, owner: user)

    get api("/v1/following", access_token: access_token.token, sort_id: "invalid")

    expect(response.status).to eq(400)
  end
end
