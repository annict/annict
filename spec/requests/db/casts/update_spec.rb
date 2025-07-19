# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/casts/:id", type: :request do
  it "ログインしていないとき、キャストを更新できずログインページにリダイレクトすること" do
    character = FactoryBot.create(:character)
    person = FactoryBot.create(:person)
    cast = FactoryBot.create(:cast)
    old_cast = cast.attributes
    cast_params = {
      character_id: character.id,
      person_id: person.id
    }

    patch "/db/casts/#{cast.id}", params: {cast: cast_params}
    cast.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")

    expect(cast.character_id).to eq(old_cast["character_id"])
    expect(cast.person_id).to eq(old_cast["person_id"])
  end

  it "編集者権限のないユーザーがログインしているとき、キャストを更新できずアクセス拒否されること" do
    character = FactoryBot.create(:character)
    person = FactoryBot.create(:person)
    user = FactoryBot.create(:registered_user)
    cast = FactoryBot.create(:cast)
    old_cast = cast.attributes
    cast_params = {
      character_id: character.id,
      person_id: person.id
    }

    login_as(user, scope: :user)

    patch "/db/casts/#{cast.id}", params: {cast: cast_params}
    cast.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(cast.character_id).to eq(old_cast["character_id"])
    expect(cast.person_id).to eq(old_cast["person_id"])
  end

  it "編集者権限のあるユーザーがログインしているとき、キャストを正常に更新できること" do
    character = FactoryBot.create(:character)
    person = FactoryBot.create(:person)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    cast = FactoryBot.create(:cast)
    old_cast = cast.attributes
    cast_params = {
      character_id: character.id,
      person_id: person.id
    }

    login_as(user, scope: :user)

    expect(cast.character_id).to eq(old_cast["character_id"])
    expect(cast.person_id).to eq(old_cast["person_id"])

    patch "/db/casts/#{cast.id}", params: {cast: cast_params}
    cast.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")

    expect(cast.character_id).to eq(character.id)
    expect(cast.person_id).to eq(person.id)
  end
end
