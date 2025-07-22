# typed: false
# frozen_string_literal: true

RSpec.describe "GET /fragment/trackable_episodes", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    get "/fragment/trackable_episodes"

    expect(response).to redirect_to("/sign_in")
  end

  it "ログインしているとき、200を返すこと" do
    user = FactoryBot.create(:user)
    login_as(user, scope: :user)

    get "/fragment/trackable_episodes"

    expect(response.status).to eq(200)
  end

  it "視聴中のアニメがあるとき、リストが表示されること" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work, :with_current_season)
    FactoryBot.create(:episode, work:)
    status = FactoryBot.create(:status, user:, work:, kind: :watching)
    FactoryBot.create(:library_entry, user:, work:, status:)
    login_as(user, scope: :user)

    get "/fragment/trackable_episodes"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "視聴中のアニメがないとき、空のリストが表示されること" do
    user = FactoryBot.create(:user)
    login_as(user, scope: :user)

    get "/fragment/trackable_episodes"

    expect(response.status).to eq(200)
    # 視聴中のアニメがないことを確認
    expect(response.body).not_to include("data-episode-item")
  end

  it "削除されたアニメは表示されないこと" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work, :with_current_season)
    deleted_work = FactoryBot.create(:work, :with_current_season, deleted_at: Time.current)
    FactoryBot.create(:episode, work:)
    FactoryBot.create(:episode, work: deleted_work)
    status = FactoryBot.create(:status, user:, work:, kind: :watching)
    deleted_status = FactoryBot.create(:status, user:, work: deleted_work, kind: :watching)
    FactoryBot.create(:library_entry, user:, work:, status:)
    FactoryBot.create(:library_entry, user:, work: deleted_work, status: deleted_status)
    login_as(user, scope: :user)

    get "/fragment/trackable_episodes"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
    expect(response.body).not_to include(deleted_work.title)
  end

  it "エピソードがないアニメは表示されないこと" do
    user = FactoryBot.create(:user)
    work_with_episodes = FactoryBot.create(:work, :with_current_season)
    work_without_episodes = FactoryBot.create(:work, :with_current_season, no_episodes: true)
    FactoryBot.create(:episode, work: work_with_episodes)
    status_with_episodes = FactoryBot.create(:status, user:, work: work_with_episodes, kind: :watching)
    status_without_episodes = FactoryBot.create(:status, user:, work: work_without_episodes, kind: :watching)
    FactoryBot.create(:library_entry, user:, work: work_with_episodes, status: status_with_episodes)
    FactoryBot.create(:library_entry, user:, work: work_without_episodes, status: status_without_episodes)
    login_as(user, scope: :user)

    get "/fragment/trackable_episodes"

    expect(response.status).to eq(200)
    expect(response.body).to include(work_with_episodes.title)
    expect(response.body).not_to include(work_without_episodes.title)
  end

  it "視聴中以外のステータスのアニメは表示されないこと" do
    user = FactoryBot.create(:user)
    watching_work = FactoryBot.create(:work, :with_current_season)
    watched_work = FactoryBot.create(:work, :with_current_season)
    FactoryBot.create(:episode, work: watching_work)
    FactoryBot.create(:episode, work: watched_work)
    watching_status = FactoryBot.create(:status, user:, work: watching_work, kind: :watching)
    watched_status = FactoryBot.create(:status, user:, work: watched_work, kind: :watched)
    FactoryBot.create(:library_entry, user:, work: watching_work, status: watching_status)
    FactoryBot.create(:library_entry, user:, work: watched_work, status: watched_status)
    login_as(user, scope: :user)

    get "/fragment/trackable_episodes"

    expect(response.status).to eq(200)
    expect(response.body).to include(watching_work.title)
    expect(response.body).not_to include(watched_work.title)
  end

  it "ページネーションが機能すること" do
    user = FactoryBot.create(:user)
    login_as(user, scope: :user)

    # 51個の視聴中アニメを作成（1ページ50件なので2ページ目が必要）
    51.times do |i|
      work = FactoryBot.create(:work, :with_current_season, title: "アニメ#{i}")
      FactoryBot.create(:episode, work:)
      status = FactoryBot.create(:status, user:, work:, kind: :watching)
      FactoryBot.create(:library_entry, user:, work:, status:, position: i)
    end

    # 1ページ目
    get "/fragment/trackable_episodes"
    expect(response.status).to eq(200)

    # 2ページ目
    get "/fragment/trackable_episodes?page=2"
    expect(response.status).to eq(200)
  end
end
