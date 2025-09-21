# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/characters/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    character = FactoryBot.create(:character, :unpublished)

    post "/db/characters/#{character.id}/publishing"
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(character.published?).to eq(false)
  end

  it "編集者権限を持たないユーザーがアクセスしたとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user)
    character = FactoryBot.create(:character, :unpublished)
    login_as(user, scope: :user)

    post "/db/characters/#{character.id}/publishing"
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(character.published?).to eq(false)
  end

  it "編集者権限を持つユーザーがアクセスしたとき、キャラクターを公開できること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    character = FactoryBot.create(:character, :unpublished)
    login_as(user, scope: :user)

    expect(character.published?).to eq(false)

    post "/db/characters/#{character.id}/publishing"
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(character.published?).to eq(true)
  end

  it "存在しないキャラクターIDを指定したとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    post "/db/characters/invalid-id/publishing"

    expect(response.status).to eq(404)
  end

  it "すでに公開済みのキャラクターを公開しようとしたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    character = FactoryBot.create(:character, :published)
    login_as(user, scope: :user)

    expect(character.published?).to eq(true)

    post "/db/characters/#{character.id}/publishing"

    expect(response.status).to eq(404)
  end
end
