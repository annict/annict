# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/episodes/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    episode = create(:episode)
    old_episode = episode.attributes
    episode_params = {
      title: "タイトルUpdated"
    }

    patch "/db/episodes/#{episode.id}", params: {episode: episode_params}
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(episode.title).to eq(old_episode["title"])
  end

  it "編集者権限を持たないユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    episode = create(:episode)
    old_episode = episode.attributes
    episode_params = {
      title: "タイトルUpdated"
    }

    login_as(user, scope: :user)

    patch "/db/episodes/#{episode.id}", params: {episode: episode_params}
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(episode.title).to eq(old_episode["title"])
  end

  it "編集者権限を持つユーザーでログインしているとき、エピソードを更新できること" do
    user = create(:registered_user, :with_editor_role)
    episode = create(:episode, title: "元のタイトル")
    episode_params = {
      title: "タイトルUpdated",
      number: "2",
      raw_number: "2.5",
      sort_number: 20
    }

    login_as(user, scope: :user)

    patch "/db/episodes/#{episode.id}", params: {episode: episode_params}
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(episode.title).to eq("タイトルUpdated")
    expect(episode.number).to eq("2")
    expect(episode.raw_number).to eq(2.5)
    expect(episode.sort_number).to eq(20)
  end

  it "編集者権限を持つユーザーでログインしているとき、前のエピソードを設定できること" do
    user = create(:registered_user, :with_editor_role)
    work = create(:work)
    prev_episode = create(:episode, work:, sort_number: 10)
    episode = create(:episode, work:, sort_number: 20)
    episode_params = {
      prev_episode_id: prev_episode.id
    }

    login_as(user, scope: :user)

    patch "/db/episodes/#{episode.id}", params: {episode: episode_params}
    episode.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(episode.prev_episode_id).to eq(prev_episode.id)
  end

  it "編集者権限を持つユーザーでログインしているとき、無効なパラメータで更新に失敗すること" do
    user = create(:registered_user, :with_editor_role)
    episode = create(:episode, sort_number: 10)
    episode_params = {
      sort_number: "invalid"
    }

    login_as(user, scope: :user)

    patch "/db/episodes/#{episode.id}", params: {episode: episode_params}
    episode.reload

    expect(response.status).to eq(422)
    expect(episode.sort_number).to eq(10)
  end

  it "編集者権限を持つユーザーでログインしているとき、許可されていないパラメータは無視されること" do
    user = create(:registered_user, :with_editor_role)
    episode = create(:episode, title: "元のタイトル")
    episode_params = {
      title: "新しいタイトル",
      work_id: 9999, # このパラメータは許可されていない
      aasm_state: "unpublished" # このパラメータも許可されていない
    }

    login_as(user, scope: :user)

    patch "/db/episodes/#{episode.id}", params: {episode: episode_params}
    episode.reload

    expect(response.status).to eq(302)
    expect(episode.title).to eq("新しいタイトル")
    expect(episode.work_id).not_to eq(9999) # work_idは変更されない
    expect(episode.aasm_state).not_to eq("unpublished") # aasm_stateも変更されない
  end

  it "存在しないエピソードIDを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    episode_params = {
      title: "タイトル"
    }

    login_as(user, scope: :user)

    patch "/db/episodes/999999", params: {episode: episode_params

    expect(response.status).to eq(404)
  end

  it "削除済みのエピソードを更新しようとしたとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    episode = create(:episode)
    episode.destroy!
    episode_params = {
      title: "タイトル"
    }

    login_as(user, scope: :user)

    patch "/db/episodes/#{episode.id

    expect(response.status).to eq(404)
  end
end
