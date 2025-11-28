# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/characters", type: :request do
  it "ユーザーがログインしていない場合、キャラクター一覧を表示すること" do
    character = create(:character)

    get "/db/characters"

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
  end

  it "ユーザーがログインしている場合、キャラクター一覧を表示すること" do
    user = create(:registered_user)
    character = create(:character)
    login_as(user, scope: :user)

    get "/db/characters"

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
  end

  it "削除されたキャラクターは表示されないこと" do
    character = create(:character)
    deleted_character = create(:character, :deleted)

    get "/db/characters"

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
    expect(response.body).not_to include(deleted_character.name)
  end

  it "ページネーションが機能すること" do
    # 101個のキャラクターを作成（1ページあたり100件なので2ページ目が必要）
    characters = create_list(:character, 101)

    get "/db/characters"

    expect(response.status).to eq(200)
    # 最新の100件が表示されること（IDの降順）
    expect(response.body).to include(characters.last.name)
    expect(response.body).not_to include(characters.first.name)
  end

  it "2ページ目のキャラクターが表示されること" do
    characters = create_list(:character, 101)

    get "/db/characters", params: {page: 2}

    expect(response.status).to eq(200)
    # 最初の1件が2ページ目に表示されること
    expect(response.body).to include(characters.first.name)
    expect(response.body).not_to include(characters.last.name)
  end

  it "シリーズ情報がプリロードされること" do
    series = create(:series, name: "テストシリーズ")
    character = create(:character, series: series)

    get "/db/characters"

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
    # シリーズ情報も表示されることを確認
    expect(response.body).to include(series.name)
  end
end
