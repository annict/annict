# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/characters/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトし、キャラクターは削除されないこと" do
    character = create(:character, :not_deleted)

    expect(Character.count).to eq(1)

    delete "/db/characters/#{character.id}"
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Character.count).to eq(1)
  end

  it "編集者権限のないユーザーがログインしているとき、アクセスできず、キャラクターは削除されないこと" do
    user = create(:registered_user)
    character = create(:character, :not_deleted)
    login_as(user, scope: :user)

    expect(Character.count).to eq(1)

    delete "/db/characters/#{character.id}"
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Character.count).to eq(1)
  end

  it "編集者権限のあるユーザーがログインしているとき、アクセスできず、キャラクターは削除されないこと" do
    user = create(:registered_user, :with_editor_role)
    character = create(:character, :not_deleted)
    login_as(user, scope: :user)

    expect(Character.count).to eq(1)

    delete "/db/characters/#{character.id}"
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Character.count).to eq(1)
  end

  it "管理者権限のあるユーザーがログインしているとき、キャラクターをソフトデリートできること" do
    user = create(:registered_user, :with_admin_role)
    character = create(:character, :not_deleted)
    login_as(user, scope: :user)

    expect(Character.count).to eq(1)

    delete "/db/characters/#{character.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
    expect(Character.count).to eq(0)
  end

  it "管理者権限のあるユーザーがログインしているとき、存在しないキャラクターIDでアクセスすると404エラーになること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    delete "/db/characters/99999999"

    expect(response.status).to eq(404)
  end

  it "管理者権限のあるユーザーがログインしているとき、既に削除されたキャラクターIDでアクセスすると404エラーになること" do
    user = create(:registered_user, :with_admin_role)
    character = create(:character, deleted_at: Time.current)
    login_as(user, scope: :user)

    delete "/db/characters/#{character.id}"

    expect(response.status).to eq(404)
  end
end
