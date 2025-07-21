# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/characters/:id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    character = FactoryBot.create(:character)

    get "/db/characters/#{character.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディター権限がないユーザーでログインしているとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user)
    character = FactoryBot.create(:character)
    login_as(user, scope: :user)

    get "/db/characters/#{character.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディター権限があるユーザーでログインしているとき、キャラクター編集フォームが表示されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    character = FactoryBot.create(:character)
    login_as(user, scope: :user)

    get "/db/characters/#{character.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
  end
end
