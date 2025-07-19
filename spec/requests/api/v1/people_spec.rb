# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/people", type: :request do
  it "パラメータなしで人物情報を取得できること" do
    access_token = create(:oauth_access_token)
    person = create(:person)

    get api("/v1/people", access_token: access_token.token)

    expect(response.status).to eq(200)

    expected_hash = {
      "id" => person.id,
      "name" => person.name,
      "name_kana" => person.name_kana,
      "name_en" => person.name_en,
      "nickname" => person.nickname,
      "nickname_en" => person.nickname_en,
      "gender_text" => person.gender_text,
      "url" => person.url,
      "url_en" => person.url_en,
      "wikipedia_url" => person.wikipedia_url,
      "wikipedia_url_en" => person.wikipedia_url_en,
      "twitter_username" => person.twitter_username,
      "twitter_username_en" => person.twitter_username_en,
      "birthday" => person.birthday&.strftime("%Y-%m-%d"),
      "blood_type" => person.blood_type,
      "height" => person.height,
      "favorite_people_count" => person.favorite_users_count,
      "casts_count" => person.casts_count,
      "staffs_count" => person.staffs_count,
      "prefecture" => nil
    }
    expect(json["people"][0]).to include(expected_hash)
    expect(json["total_count"]).to eq(1)
    expect(json["next_page"]).to eq(nil)
    expect(json["prev_page"]).to eq(nil)
  end

  it "アクセストークンが提供されていない場合、401エラーを返すこと" do
    get api("/v1/people")
    expect(response.status).to eq(401)
  end

  it "無効なアクセストークンが提供された場合、401エラーを返すこと" do
    get api("/v1/people", access_token: "invalid_token")
    expect(response.status).to eq(401)
  end

  it "filter_idsで人物をフィルタリングできること" do
    access_token = create(:oauth_access_token)
    person1 = create(:person)
    create(:person)
    person3 = create(:person)

    get api("/v1/people", access_token: access_token.token, filter_ids: "#{person1.id},#{person3.id}")

    expect(response.status).to eq(200)
    expect(json["people"].size).to eq(2)
    expect(json["people"].pluck("id")).to contain_exactly(person1.id, person3.id)
    expect(json["total_count"]).to eq(2)
  end

  it "filter_nameで人物名を検索できること" do
    access_token = create(:oauth_access_token)
    person1 = create(:person, name: "テスト声優")
    create(:person, name: "別の人")

    get api("/v1/people", access_token: access_token.token, filter_name: "テスト")

    expect(response.status).to eq(200)
    expect(json["people"].size).to eq(1)
    expect(json["people"][0]["id"]).to eq(person1.id)
    expect(json["people"][0]["name"]).to eq("テスト声優")
  end

  it "sort_idで人物を昇順でソートできること" do
    access_token = create(:oauth_access_token)
    create(:person)
    create(:person)
    create(:person)

    get api("/v1/people", access_token: access_token.token, sort_id: "asc")

    expect(response.status).to eq(200)
    expect(json["people"].size).to eq(3)
    ids = json["people"].pluck("id")
    expect(ids).to eq(ids.sort)
  end

  it "sort_idで人物を降順でソートできること" do
    access_token = create(:oauth_access_token)
    create(:person)
    create(:person)
    create(:person)

    get api("/v1/people", access_token: access_token.token, sort_id: "desc")

    expect(response.status).to eq(200)
    expect(json["people"].size).to eq(3)
    ids = json["people"].pluck("id")
    expect(ids).to eq(ids.sort.reverse)
  end

  it "pageとper_pageでページネーションができること" do
    access_token = create(:oauth_access_token)
    create_list(:person, 5)

    get api("/v1/people", access_token: access_token.token, page: 1, per_page: 2)

    expect(response.status).to eq(200)
    expect(json["people"].size).to eq(2)
    expect(json["total_count"]).to eq(5)
    expect(json["next_page"]).to eq(2)
    expect(json["prev_page"]).to eq(nil)
  end

  it "削除された人物は表示されないこと" do
    access_token = create(:oauth_access_token)
    person1 = create(:person)
    create(:person, deleted_at: Time.current)

    get api("/v1/people", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["people"].size).to eq(1)
    expect(json["people"][0]["id"]).to eq(person1.id)
    expect(json["total_count"]).to eq(1)
  end

  it "未公開の人物は表示されないこと" do
    access_token = create(:oauth_access_token)
    person1 = create(:person)
    create(:person, unpublished_at: Time.current)

    get api("/v1/people", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["people"].size).to eq(1)
    expect(json["people"][0]["id"]).to eq(person1.id)
    expect(json["total_count"]).to eq(1)
  end

  it "人物が存在しない場合、空の配列を返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/people", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["people"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end

  it "per_pageが上限を超える場合、400エラーを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/people", access_token: access_token.token, per_page: 51)

    expect(response.status).to eq(400)
  end

  it "per_pageが0の場合、400エラーを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/people", access_token: access_token.token, per_page: 0)

    expect(response.status).to eq(400)
  end

  it "pageが0の場合、400エラーを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/people", access_token: access_token.token, page: 0)

    expect(response.status).to eq(400)
  end

  it "sort_idが無効な値の場合、400エラーを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/people", access_token: access_token.token, sort_id: "invalid")

    expect(response.status).to eq(400)
  end

  it "存在しない人物名で検索した場合、空の結果を返すこと" do
    access_token = create(:oauth_access_token)
    create(:person, name: "テスト人物")

    get api("/v1/people", access_token: access_token.token, filter_name: "存在しない名前")

    expect(response.status).to eq(200)
    expect(json["people"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end

  it "存在しないIDでフィルタリングした場合、空の結果を返すこと" do
    access_token = create(:oauth_access_token)
    create(:person)

    get api("/v1/people", access_token: access_token.token, filter_ids: "999999")

    expect(response.status).to eq(200)
    expect(json["people"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end
end
