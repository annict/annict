# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/episodes/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    episode = create(:episode, :not_deleted)

    expect(Episode.count).to eq(1)

    delete "/db/episodes/#{episode.id}"
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Episode.count).to eq(1)
  end

  it "編集者権限を持たないユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    episode = create(:episode, :not_deleted)
    login_as(user, scope: :user)

    expect(Episode.count).to eq(1)

    delete "/db/episodes/#{episode.id}"
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Episode.count).to eq(1)
  end

  it "編集者権限を持つユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    episode = create(:episode, :not_deleted)
    login_as(user, scope: :user)

    expect(Episode.count).to eq(1)

    delete "/db/episodes/#{episode.id}"
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Episode.count).to eq(1)
  end

  it "管理者権限を持つユーザーでログインしているとき、エピソードを論理削除できること" do
    user = create(:registered_user, :with_admin_role)
    episode = create(:episode, :not_deleted)
    login_as(user, scope: :user)

    expect(Episode.count).to eq(1)

    delete "/db/episodes/#{episode.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
    expect(Episode.count).to eq(0)
  end

  it "存在しないエピソードIDを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    expect { delete "/db/episodes/non-existent-id" }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済みのエピソードを削除しようとしたとき、404エラーになること" do
    user = create(:registered_user, :with_admin_role)
    episode = create(:episode, deleted_at: Time.current)
    login_as(user, scope: :user)

    expect { delete "/db/episodes/#{episode.id}" }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
