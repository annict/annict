# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/characters", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    series = create(:series)
    character_params = {
      rows: "かぐや姫,かぐやひめ,#{series.name}"
    }

    post "/db/characters", params: {deprecated_db_character_rows_form: character_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")

    expect(Character.all.size).to eq(0)
  end

  it "エディター権限がないユーザーでログインしているとき、アクセスできないこと" do
    series = create(:series)
    user = create(:registered_user)
    character_params = {
      rows: "かぐや姫,かぐやひめ,#{series.name}"
    }

    login_as(user, scope: :user)

    post "/db/characters", params: {deprecated_db_character_rows_form: character_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(Character.all.size).to eq(0)
  end

  it "エディター権限があるユーザーでログインしているとき、キャラクターを作成できること" do
    series = create(:series)
    user = create(:registered_user, :with_editor_role)
    character_params = {
      rows: "かぐや姫,かぐやひめ,#{series.name}"
    }

    login_as(user, scope: :user)

    expect(Character.all.size).to eq(0)

    post "/db/characters", params: {deprecated_db_character_rows_form: character_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")

    expect(Character.all.size).to eq(1)
    character = Character.first

    expect(character.name).to eq("かぐや姫")
    expect(character.name_kana).to eq("かぐやひめ")
    expect(character.series_id).to eq(series.id)
  end
end
