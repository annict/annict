# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/characters/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    character = create(:character)
    old_character = character.attributes
    character_params = {
      name: "かぐや姫"
    }

    patch "/db/characters/#{character.id}", params: {character: character_params}
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")

    expect(character.name).to eq(old_character["name"])
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    character = create(:character)
    old_character = character.attributes
    character_params = {
      name: "かぐや姫"
    }

    login_as(user, scope: :user)

    patch "/db/characters/#{character.id}", params: {character: character_params}
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(character.name).to eq(old_character["name"])
  end

  it "編集者権限を持つユーザーがログインしているとき、キャラクターを更新できること" do
    user = create(:registered_user, :with_editor_role)
    character = create(:character)
    old_character = character.attributes
    character_params = {
      name: "かぐや姫"
    }

    login_as(user, scope: :user)

    expect(character.name).to eq(old_character["name"])

    patch "/db/characters/#{character.id}", params: {character: character_params}
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")

    expect(character.name).to eq("かぐや姫")
  end

  it "編集者権限を持つユーザーがログインしているとき、複数のパラメータを更新できること" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series)
    character = create(:character)
    character_params = {
      name: "竹取物語のかぐや姫",
      name_kana: "たけとりものがたりのかぐやひめ",
      name_en: "Princess Kaguya",
      series_id: series.id,
      nickname: "かぐや",
      nickname_en: "Kaguya",
      birthday: "8月15日",
      birthday_en: "August 15",
      age: "16",
      age_en: "16",
      blood_type: "AB",
      blood_type_en: "AB",
      height: "165cm",
      height_en: "165cm",
      weight: "48kg",
      weight_en: "48kg",
      nationality: "日本",
      nationality_en: "Japan",
      occupation: "姫",
      occupation_en: "Princess",
      description: "月の世界から来た美しい姫。",
      description_en: "A beautiful princess from the moon.",
      description_source: "竹取物語",
      description_source_en: "The Tale of the Bamboo Cutter"
    }

    login_as(user, scope: :user)

    patch "/db/characters/#{character.id}", params: {character: character_params}
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")

    expect(character.name).to eq("竹取物語のかぐや姫")
    expect(character.name_kana).to eq("たけとりものがたりのかぐやひめ")
    expect(character.name_en).to eq("Princess Kaguya")
    expect(character.series_id).to eq(series.id)
    expect(character.nickname).to eq("かぐや")
    expect(character.nickname_en).to eq("Kaguya")
    expect(character.birthday).to eq("8月15日")
    expect(character.birthday_en).to eq("August 15")
    expect(character.age).to eq("16")
    expect(character.age_en).to eq("16")
    expect(character.blood_type).to eq("AB")
    expect(character.blood_type_en).to eq("AB")
    expect(character.height).to eq("165cm")
    expect(character.height_en).to eq("165cm")
    expect(character.weight).to eq("48kg")
    expect(character.weight_en).to eq("48kg")
    expect(character.nationality).to eq("日本")
    expect(character.nationality_en).to eq("Japan")
    expect(character.occupation).to eq("姫")
    expect(character.occupation_en).to eq("Princess")
    expect(character.description).to eq("月の世界から来た美しい姫。")
    expect(character.description_en).to eq("A beautiful princess from the moon.")
    expect(character.description_source).to eq("竹取物語")
    expect(character.description_source_en).to eq("The Tale of the Bamboo Cutter")
  end

  it "編集者権限を持つユーザーがログインしているとき、名前が空の場合バリデーションエラーになること" do
    user = create(:registered_user, :with_editor_role)
    character = create(:character)
    old_name = character.name
    character_params = {
      name: ""
    }

    login_as(user, scope: :user)

    patch "/db/characters/#{character.id}", params: {character: character_params}
    character.reload

    expect(response.status).to eq(422)
    expect(character.name).to eq(old_name)
  end

  it "編集者権限を持つユーザーがログインしているとき、series_idが空の場合バリデーションエラーになること" do
    user = create(:registered_user, :with_editor_role)
    character = create(:character)
    old_series_id = character.series_id
    character_params = {
      series_id: ""
    }

    login_as(user, scope: :user)

    patch "/db/characters/#{character.id}", params: {character: character_params}
    character.reload

    expect(response.status).to eq(422)
    expect(character.series_id).to eq(old_series_id)
  end

  it "編集者権限を持つユーザーがログインしているとき、削除されたキャラクターは見つからないこと" do
    user = create(:registered_user, :with_editor_role)
    character = create(:character, deleted_at: Time.current)
    character_params = {
      name: "更新されない名前"
    }

    login_as(user, scope: :user)

    patch "/db/characters/#{character.id

    expect(response.status).to eq(404)
  end

  it "編集者権限を持つユーザーがログインしているとき、存在しないキャラクターIDの場合見つからないこと" do
    user = create(:registered_user, :with_editor_role)
    character_params = {
      name: "更新されない名前"
    }

    login_as(user, scope: :user)

    patch "/db/characters/9999999", params: {character: character_params

    expect(response.status).to eq(404)
  end

  it "編集者権限を持つユーザーがログインしているとき、descriptionとdescription_sourceが両方設定されること" do
    user = create(:registered_user, :with_editor_role)
    character = create(:character)
    character_params = {
      description: "テストの説明文",
      description_source: "テスト出典"
    }

    login_as(user, scope: :user)

    patch "/db/characters/#{character.id}", params: {character: character_params}
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(character.description).to eq("テストの説明文")
    expect(character.description_source).to eq("テスト出典")
  end

  it "編集者権限を持つユーザーがログインしているとき、アクティビティが作成されること" do
    user = create(:registered_user, :with_editor_role)
    character = create(:character)
    character_params = {
      name: "更新後の名前"
    }

    login_as(user, scope: :user)

    expect {
      patch "/db/characters/#{character.id}", params: {character: character_params}
    }.to change(DbActivity, :count).by(1)

    activity = DbActivity.last
    expect(activity.trackable).to eq(character)
    expect(activity.action).to eq("characters.update")
  end
end
