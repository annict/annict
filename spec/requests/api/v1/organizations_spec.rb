# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/organizations", type: :request do
  it "パラメータなしで組織情報を取得できること" do
    access_token = create(:oauth_access_token)
    organization = create(:organization)

    get api("/v1/organizations", access_token: access_token.token)

    expect(response.status).to eq(200)

    expected_hash = {
      "id" => organization.id,
      "name" => organization.name,
      "name_kana" => organization.name_kana,
      "name_en" => organization.name_en,
      "url" => organization.url,
      "url_en" => organization.url_en,
      "wikipedia_url" => organization.wikipedia_url,
      "wikipedia_url_en" => organization.wikipedia_url_en,
      "twitter_username" => organization.twitter_username,
      "twitter_username_en" => organization.twitter_username_en,
      "favorite_organizations_count" => organization.favorite_users_count,
      "staffs_count" => organization.staffs_count
    }
    expect(json["organizations"][0]).to include(expected_hash)
    expect(json["total_count"]).to eq(1)
    expect(json["next_page"]).to eq(nil)
    expect(json["prev_page"]).to eq(nil)
  end

  it "アクセストークンが提供されていない場合、401エラーを返すこと" do
    get api("/v1/organizations")
    expect(response.status).to eq(401)
  end

  it "無効なアクセストークンが提供された場合、401エラーを返すこと" do
    get api("/v1/organizations", access_token: "invalid_token")
    expect(response.status).to eq(401)
  end

  it "filter_idsで組織をフィルタリングできること" do
    access_token = create(:oauth_access_token)
    organization1 = create(:organization)
    create(:organization)
    organization3 = create(:organization)

    get api("/v1/organizations", access_token: access_token.token, filter_ids: "#{organization1.id},#{organization3.id}")

    expect(response.status).to eq(200)
    expect(json["organizations"].size).to eq(2)
    expect(json["organizations"].pluck("id")).to contain_exactly(organization1.id, organization3.id)
    expect(json["total_count"]).to eq(2)
  end

  it "filter_nameで組織名を検索できること" do
    access_token = create(:oauth_access_token)
    organization1 = create(:organization, name: "テスト制作会社")
    create(:organization, name: "別の会社")

    get api("/v1/organizations", access_token: access_token.token, filter_name: "テスト")

    expect(response.status).to eq(200)
    expect(json["organizations"].size).to eq(1)
    expect(json["organizations"][0]["id"]).to eq(organization1.id)
    expect(json["organizations"][0]["name"]).to eq("テスト制作会社")
  end

  it "sort_idで組織を昇順でソートできること" do
    access_token = create(:oauth_access_token)
    create(:organization)
    create(:organization)
    create(:organization)

    get api("/v1/organizations", access_token: access_token.token, sort_id: "asc")

    expect(response.status).to eq(200)
    expect(json["organizations"].size).to eq(3)
    ids = json["organizations"].pluck("id")
    expect(ids).to eq(ids.sort)
  end

  it "sort_idで組織を降順でソートできること" do
    access_token = create(:oauth_access_token)
    create(:organization)
    create(:organization)
    create(:organization)

    get api("/v1/organizations", access_token: access_token.token, sort_id: "desc")

    expect(response.status).to eq(200)
    expect(json["organizations"].size).to eq(3)
    ids = json["organizations"].pluck("id")
    expect(ids).to eq(ids.sort.reverse)
  end

  it "pageとper_pageでページネーションができること" do
    access_token = create(:oauth_access_token)
    create_list(:organization, 5)

    get api("/v1/organizations", access_token: access_token.token, page: 1, per_page: 2)

    expect(response.status).to eq(200)
    expect(json["organizations"].size).to eq(2)
    expect(json["total_count"]).to eq(5)
    expect(json["next_page"]).to eq(2)
    expect(json["prev_page"]).to eq(nil)
  end

  it "削除された組織は表示されないこと" do
    access_token = create(:oauth_access_token)
    organization1 = create(:organization)
    create(:organization, deleted_at: Time.current)

    get api("/v1/organizations", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["organizations"].size).to eq(1)
    expect(json["organizations"][0]["id"]).to eq(organization1.id)
    expect(json["total_count"]).to eq(1)
  end

  it "組織が存在しない場合、空の配列を返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/organizations", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["organizations"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end

  it "per_pageが上限を超える場合、400エラーを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/organizations", access_token: access_token.token, per_page: 51)

    expect(response.status).to eq(400)
  end

  it "per_pageが0の場合、400エラーを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/organizations", access_token: access_token.token, per_page: 0)

    expect(response.status).to eq(400)
  end

  it "pageが0の場合、400エラーを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/organizations", access_token: access_token.token, page: 0)

    expect(response.status).to eq(400)
  end

  it "sort_idが無効な値の場合、400エラーを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/organizations", access_token: access_token.token, sort_id: "invalid")

    expect(response.status).to eq(400)
  end

  it "存在しない組織名で検索した場合、空の結果を返すこと" do
    access_token = create(:oauth_access_token)
    create(:organization, name: "テスト組織")

    get api("/v1/organizations", access_token: access_token.token, filter_name: "存在しない名前")

    expect(response.status).to eq(200)
    expect(json["organizations"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end

  it "存在しないIDでフィルタリングした場合、空の結果を返すこと" do
    access_token = create(:oauth_access_token)
    create(:organization)

    get api("/v1/organizations", access_token: access_token.token, filter_ids: "999999")

    expect(response.status).to eq(200)
    expect(json["organizations"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end
end
