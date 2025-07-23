# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/episodes/:id/edit", type: :request do
  it "ログインしていない場合、ログインページにリダイレクトされること" do
    episode = create(:episode)

    get "/db/episodes/#{episode.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限を持たないユーザーがログインしている場合、アクセスできないこと" do
    user = create(:registered_user)
    episode = create(:episode)
    login_as(user, scope: :user)

    get "/db/episodes/#{episode.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限を持つユーザーがログインしている場合、エピソード編集フォームが表示されること" do
    user = create(:registered_user, :with_editor_role)
    episode = create(:episode)
    login_as(user, scope: :user)

    get "/db/episodes/#{episode.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(episode.title)
  end

  it "管理者権限を持つユーザーがログインしている場合、エピソード編集フォームが表示されること" do
    user = create(:registered_user, :with_admin_role)
    episode = create(:episode)
    login_as(user, scope: :user)

    get "/db/episodes/#{episode.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(episode.title)
  end

  it "削除されたエピソードの場合、404エラーが発生すること" do
    user = create(:registered_user, :with_editor_role)
    episode = create(:episode, deleted_at: Time.current)
    login_as(user, scope: :user)

    expect {
      get "/db/episodes/#{episode.id}/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないエピソードIDの場合、404エラーが発生すること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect {
      get "/db/episodes/999999/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
