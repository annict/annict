# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/skipped_episodes", type: :request do
  it "未認証ユーザーの場合、リダイレクトされること" do
    episode = FactoryBot.create(:episode)

    post "/api/internal/skipped_episodes", params: {
      episode_id: episode.id
    }

    expect(response.status).to eq(302)
  end

  it "存在しないエピソードIDの場合、404を返すこと" do
    user = FactoryBot.create(:user, :with_profile)

    login_as(user, scope: :user)

    expect {
      post "/api/internal/skipped_episodes", params: {
        episode_id: "nonexistent"
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたエピソードの場合、404を返すこと" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode, deleted_at: Time.zone.now)

    login_as(user, scope: :user)

    expect {
      post "/api/internal/skipped_episodes", params: {
        episode_id: episode.id
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "有効なパラメータでエピソードをスキップできること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)

    login_as(user, scope: :user)
    post "/api/internal/skipped_episodes", params: {
      episode_id: episode.id
    }

    expect(response.status).to eq(204)

    # library_entryが作成されていることを確認
    library_entry = user.library_entries.find_by(work_id: episode.work_id)
    expect(library_entry).to be_present

    # エピソードがwatched_episode_idsに追加されていることを確認
    expect(library_entry.watched_episode_ids).to include(episode.id)
  end

  it "既存のlibrary_entryがある場合も正常に動作すること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)
    work = episode.work

    # 既存のlibrary_entryを作成
    existing_library_entry = FactoryBot.create(:library_entry, user: user, work: work)

    login_as(user, scope: :user)
    post "/api/internal/skipped_episodes", params: {
      episode_id: episode.id
    }

    expect(response.status).to eq(204)

    # 既存のlibrary_entryが使われていることを確認
    library_entries = user.library_entries.where(work_id: work.id)
    expect(library_entries.count).to eq(1)
    expect(library_entries.first).to eq(existing_library_entry)

    # エピソードがwatched_episode_idsに追加されていることを確認
    library_entry = library_entries.first
    expect(library_entry.watched_episode_ids).to include(episode.id)
  end

  it "既にスキップ済みのエピソードでも正常に処理されること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)

    # 既にスキップ済みの状態を作成
    library_entry = FactoryBot.create(:library_entry,
      user: user,
      work: episode.work,
      watched_episode_ids: [episode.id])

    login_as(user, scope: :user)
    post "/api/internal/skipped_episodes", params: {
      episode_id: episode.id
    }

    expect(response.status).to eq(204)

    # エピソードが依然としてwatched_episode_idsに含まれていることを確認
    library_entry.reload
    expect(library_entry.watched_episode_ids).to include(episode.id)
    expect(library_entry.watched_episode_ids.count(episode.id)).to eq(1) # 重複していないことを確認
  end

  it "他のエピソードが視聴済みの状態でも新しいエピソードをスキップできること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    episode1 = FactoryBot.create(:episode, work: work)
    episode2 = FactoryBot.create(:episode, work: work)

    # 既に別のエピソードが視聴済みの状態を作成
    library_entry = FactoryBot.create(:library_entry,
      user: user,
      work: work,
      watched_episode_ids: [episode1.id])

    login_as(user, scope: :user)
    post "/api/internal/skipped_episodes", params: {
      episode_id: episode2.id
    }

    expect(response.status).to eq(204)

    # 両方のエピソードがwatched_episode_idsに含まれていることを確認
    library_entry.reload
    expect(library_entry.watched_episode_ids).to include(episode1.id, episode2.id)
  end
end
