# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/characters/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    character = create(:character, :published)

    delete "/db/characters/#{character.id}/publishing"
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(character.published?).to eq(true)
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    character = create(:character, :published)
    login_as(user, scope: :user)

    delete "/db/characters/#{character.id}/publishing"
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(character.published?).to eq(true)
  end

  it "編集者権限を持つユーザーがログインしているとき、キャラクターを非公開にできること" do
    user = create(:registered_user, :with_editor_role)
    character = create(:character, :published)
    login_as(user, scope: :user)

    expect(character.published?).to eq(true)

    delete "/db/characters/#{character.id}/publishing"
    character.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(character.published?).to eq(false)
  end

  it "存在しないキャラクターIDを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    delete "/db/characters/99999/publishing"

    expect(response).to have_http_status(404)
  end

  it "未公開のキャラクターを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    character = create(:character, :unpublished)
    login_as(user, scope: :user)

    expect(character.published?).to eq(false)

    delete "/db/characters/#{character.id}/publishing"

    expect(response).to have_http_status(404)
  end

  it "削除済みのキャラクターを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    character = create(:character, :published)
    character.destroy_in_batches
    login_as(user, scope: :user)

    delete "/db/characters/#{character.id}/publishing"

    expect(response).to have_http_status(404)
  end
end
