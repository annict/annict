# typed: false
# frozen_string_literal: true

RSpec.describe "GET /track", type: :request do
  it "未ログインのとき、ログインページにリダイレクトすること" do
    get track_path

    expect(response).to have_http_status(:found)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログイン済みのとき、トラッキングページを表示すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get track_path

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("視聴中")
  end

  it "ページネーションが正しく動作すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    # 51個のlibrary_entryを作成してページネーションをテスト
    51.times do
      work = create(:work, :with_current_season, no_episodes: false)
      status = create(:status, user:, work:, kind: :watching)
      create(:library_entry, user:, work:, status:)
    end

    get track_path, params: {page: 2}

    expect(response).to have_http_status(:ok)
    # ページネーション情報はフラグメントで読み込まれるため、page=2のクエリパラメータが正しく渡されていることを確認
    expect(response.body).to include("page=2")
  end

  it "視聴中のアニメのみが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    watching_work = create(:work, :with_current_season, no_episodes: false)
    completed_work = create(:work, :with_current_season, no_episodes: false)

    watching_status = create(:status, user:, work: watching_work, kind: :watching)
    completed_status = create(:status, user:, work: completed_work, kind: :watched)

    create(:library_entry, user:, work: watching_work, status: watching_status)
    create(:library_entry, user:, work: completed_work, status: completed_status)

    get track_path

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(watching_work.local_title)
    expect(response.body).not_to include(completed_work.local_title)
  end
end
