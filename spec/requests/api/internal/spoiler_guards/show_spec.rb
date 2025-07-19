# typed: false
# frozen_string_literal: true

RSpec.describe "GET /api/internal/spoiler_guard", type: :request do
  it "未ログインの場合、ログイン状態がfalseで空の配列を返すこと" do
    logout(:user)
    get internal_api_spoiler_guard_path

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["is_signed_in"]).to eq(false)
    expect(json["episode_ids"]).to eq([])
    expect(json["work_ids"]).to eq([])
  end

  it "ログイン済みの場合、ユーザーの視聴情報を返すこと" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    # 視聴済み作品
    work1 = create(:work)
    work2 = create(:work)
    create(:work_record, user:, work: work1)
    create(:work_record, user:, work: work2)

    # 視聴済みエピソード
    episode1 = create(:episode)
    episode2 = create(:episode)
    create(:episode_record, user:, episode: episode1)
    create(:episode_record, user:, episode: episode2)

    # 視聴完了作品
    work3 = create(:work)
    status_watched = create(:status, user:, work: work3, kind: :watched)
    create(:library_entry, user:, work: work3, status: status_watched)

    # ライブラリに追加した作品
    work4 = create(:work)
    status_watching = create(:status, user:, work: work4, kind: :watching)
    create(:library_entry, user:, work: work4, status: status_watching)

    get internal_api_spoiler_guard_path

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["is_signed_in"]).to eq(true)
    expect(json["hide_record_body"]).to eq(user.hide_record_body?)
    expect(json["watched_work_ids"]).to contain_exactly(work1.id, work2.id, work3.id)
    expect(json["work_ids_in_library"]).to contain_exactly(work3.id, work4.id)
    expect(json["tracked_episode_ids"]).to contain_exactly(episode1.id, episode2.id)
  end

  it "削除されたレコードは含まれないこと" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    work = create(:work)
    episode = create(:episode)

    # 削除されたレコード
    create(:work_record, user:, work:, deleted_at: Time.current)
    create(:episode_record, user:, episode:, deleted_at: Time.current)

    # 有効なレコード
    work2 = create(:work)
    episode2 = create(:episode)
    create(:work_record, user:, work: work2)
    create(:episode_record, user:, episode: episode2)

    get internal_api_spoiler_guard_path

    json = JSON.parse(response.body)
    expect(json["watched_work_ids"]).to contain_exactly(work2.id)
    expect(json["tracked_episode_ids"]).to contain_exactly(episode2.id)
  end

  it "hide_record_bodyが有効な場合、trueを返すこと" do
    user = create(:registered_user)
    user.setting.update!(hide_record_body: true)
    login_as(user, scope: :user)

    get internal_api_spoiler_guard_path

    json = JSON.parse(response.body)
    expect(json["hide_record_body"]).to eq(true)
  end
end
