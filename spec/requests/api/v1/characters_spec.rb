# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/characters", type: :request do
  it "パラメータなしでキャラクター情報を取得できること" do
    access_token = create(:oauth_access_token)
    character = create(:character)

    get api("/v1/characters", access_token: access_token.token)

    expect(response.status).to eq(200)

    expected_hash = {
      "id" => character.id,
      "name" => character.name,
      "name_kana" => character.name_kana,
      "name_en" => character.name_en,
      "nickname" => character.nickname,
      "nickname_en" => character.nickname_en,
      "birthday" => character.birthday,
      "birthday_en" => character.birthday_en,
      "age" => character.age,
      "age_en" => character.age_en,
      "blood_type" => character.blood_type,
      "blood_type_en" => character.blood_type_en,
      "height" => character.height,
      "height_en" => character.height_en,
      "weight" => character.weight,
      "weight_en" => character.weight_en,
      "nationality" => character.nationality,
      "nationality_en" => character.nationality_en,
      "occupation" => character.occupation,
      "occupation_en" => character.occupation_en,
      "description" => character.description,
      "description_en" => character.description_en,
      "description_source" => character.description_source,
      "description_source_en" => character.description_source_en,
      "favorite_characters_count" => character.favorite_users_count,
      "series" => {
        "id" => character.series.id,
        "name" => character.series.name,
        "name_en" => character.series.name_en,
        "name_ro" => character.series.name_ro
      }
    }
    expect(json["characters"][0]).to include(expected_hash)
    expect(json["total_count"]).to eq(1)
    expect(json["next_page"]).to eq(nil)
    expect(json["prev_page"]).to eq(nil)
  end

  it "filter_idsでキャラクターをフィルタリングできること" do
    access_token = create(:oauth_access_token)
    character1 = create(:character)
    create(:character)
    character3 = create(:character)

    get api("/v1/characters", access_token: access_token.token, filter_ids: "#{character1.id},#{character3.id}")

    expect(response.status).to eq(200)
    expect(json["characters"].size).to eq(2)
    expect(json["characters"].pluck("id")).to contain_exactly(character1.id, character3.id)
    expect(json["total_count"]).to eq(2)
  end

  it "filter_nameでキャラクター名を検索できること" do
    access_token = create(:oauth_access_token)
    character1 = create(:character, name: "テストキャラクター1")
    create(:character, name: "別のキャラクター")

    get api("/v1/characters", access_token: access_token.token, filter_name: "テスト")

    expect(response.status).to eq(200)
    expect(json["characters"].size).to eq(1)
    expect(json["characters"][0]["id"]).to eq(character1.id)
    expect(json["characters"][0]["name"]).to eq("テストキャラクター1")
  end

  it "sort_idでキャラクターを昇順でソートできること" do
    access_token = create(:oauth_access_token)
    create(:character)
    create(:character)
    create(:character)

    get api("/v1/characters", access_token: access_token.token, sort_id: "asc")

    expect(response.status).to eq(200)
    expect(json["characters"].size).to eq(3)
    ids = json["characters"].pluck("id")
    expect(ids).to eq(ids.sort)
  end

  it "sort_idでキャラクターを降順でソートできること" do
    access_token = create(:oauth_access_token)
    create(:character)
    create(:character)
    create(:character)

    get api("/v1/characters", access_token: access_token.token, sort_id: "desc")

    expect(response.status).to eq(200)
    expect(json["characters"].size).to eq(3)
    ids = json["characters"].pluck("id")
    expect(ids).to eq(ids.sort.reverse)
  end

  it "pageとper_pageでページネーションができること" do
    access_token = create(:oauth_access_token)
    create_list(:character, 5)

    get api("/v1/characters", access_token: access_token.token, page: 1, per_page: 2)

    expect(response.status).to eq(200)
    expect(json["characters"].size).to eq(2)
    expect(json["total_count"]).to eq(5)
    expect(json["next_page"]).to eq(2)
    expect(json["prev_page"]).to eq(nil)
  end

  it "2ページ目を取得できること" do
    access_token = create(:oauth_access_token)
    create_list(:character, 5)

    get api("/v1/characters", access_token: access_token.token, page: 2, per_page: 2)

    expect(response.status).to eq(200)
    expect(json["characters"].size).to eq(2)
    expect(json["total_count"]).to eq(5)
    expect(json["next_page"]).to eq(3)
    expect(json["prev_page"]).to eq(1)
  end

  it "削除されたキャラクターは表示されないこと" do
    access_token = create(:oauth_access_token)
    character1 = create(:character)
    create(:character, deleted_at: Time.current)

    get api("/v1/characters", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["characters"].size).to eq(1)
    expect(json["characters"][0]["id"]).to eq(character1.id)
    expect(json["total_count"]).to eq(1)
  end

  it "未公開のキャラクターは表示されないこと" do
    access_token = create(:oauth_access_token)
    character1 = create(:character)
    create(:character, unpublished_at: Time.current)

    get api("/v1/characters", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["characters"].size).to eq(1)
    expect(json["characters"][0]["id"]).to eq(character1.id)
    expect(json["total_count"]).to eq(1)
  end

  it "アクセストークンが提供されていない場合、401エラーを返すこと" do
    get api("/v1/characters")

    expect(response.status).to eq(401)
  end

  it "無効なアクセストークンが提供された場合、401エラーを返すこと" do
    get api("/v1/characters", access_token: "invalid_token")

    expect(response.status).to eq(401)
  end

  it "per_pageが上限を超える場合、400エラーを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/characters", access_token: access_token.token, per_page: 51)

    expect(response.status).to eq(400)
  end

  it "per_pageが0の場合、400エラーを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/characters", access_token: access_token.token, per_page: 0)

    expect(response.status).to eq(400)
  end

  it "pageが0の場合、400エラーを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/characters", access_token: access_token.token, page: 0)

    expect(response.status).to eq(400)
  end

  it "sort_idが無効な値の場合、400エラーを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/characters", access_token: access_token.token, sort_id: "invalid")

    expect(response.status).to eq(400)
  end

  it "キャラクターが存在しない場合、空の配列を返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/characters", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["characters"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end

  it "存在しないキャラクター名で検索した場合、空の結果を返すこと" do
    access_token = create(:oauth_access_token)
    create(:character, name: "テストキャラクター")

    get api("/v1/characters", access_token: access_token.token, filter_name: "存在しない名前")

    expect(response.status).to eq(200)
    expect(json["characters"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end

  it "存在しないIDでフィルタリングした場合、空の結果を返すこと" do
    access_token = create(:oauth_access_token)
    create(:character)

    get api("/v1/characters", access_token: access_token.token, filter_ids: "999999")

    expect(response.status).to eq(200)
    expect(json["characters"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end
end
