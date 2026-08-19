# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/episodes/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    episode = FactoryBot.create(:episode, :published)

    delete "/db/episodes/#{episode.id}/publishing"
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(episode.published?).to eq(true)
  end

  it "編集者権限を持たないユーザーでログインしているとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user)
    episode = FactoryBot.create(:episode, :published)
    login_as(user, scope: :user)

    delete "/db/episodes/#{episode.id}/publishing"
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(episode.published?).to eq(true)
  end

  it "編集者権限を持つユーザーでログインしているとき、エピソードを非公開にできること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    episode = FactoryBot.create(:episode, :published)
    login_as(user, scope: :user)

    expect(episode.published?).to eq(true)

    delete "/db/episodes/#{episode.id}/publishing"
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(episode.published?).to eq(false)
  end

  it "存在しないエピソードIDを指定したとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect {
      delete "/db/episodes/nonexistent-id/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "既に非公開のエピソードを指定したとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    episode = FactoryBot.create(:episode, :unpublished)
    login_as(user, scope: :user)

    expect {
      delete "/db/episodes/#{episode.id}/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたエピソードを指定したとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    episode = FactoryBot.create(:episode, :published, deleted_at: Time.current)
    login_as(user, scope: :user)

    expect {
      delete "/db/episodes/#{episode.id}/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
