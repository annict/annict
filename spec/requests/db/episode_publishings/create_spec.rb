# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/episodes/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    episode = create(:episode, :unpublished)

    post "/db/episodes/#{episode.id}/publishing"
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(episode.published?).to eq(false)
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    episode = create(:episode, :unpublished)
    login_as(user, scope: :user)

    post "/db/episodes/#{episode.id}/publishing"
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(episode.published?).to eq(false)
  end

  it "編集者権限を持つユーザーがログインしているとき、エピソードを公開できること" do
    user = create(:registered_user, :with_editor_role)
    episode = create(:episode, :unpublished)
    login_as(user, scope: :user)

    expect(episode.published?).to eq(false)

    post "/db/episodes/#{episode.id}/publishing"
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(episode.published?).to eq(true)
  end

  it "存在しないエピソードIDを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect {
      post "/db/episodes/99999999/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済みのエピソードを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    episode = create(:episode, :unpublished, deleted_at: Time.current)
    login_as(user, scope: :user)

    expect {
      post "/db/episodes/#{episode.id}/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "既に公開済みのエピソードを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    episode = create(:episode, :published)
    login_as(user, scope: :user)

    expect {
      post "/db/episodes/#{episode.id}/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
