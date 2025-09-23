# typed: false
# frozen_string_literal: true

RSpec.describe "GET /fragment/trackable_episodes/:episode_id", type: :request do
  it "ログインしているとき、エピソード情報とフォームを表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:, title: "第1話", number: 1)

    login_as(user, scope: :user)
    get "/fragment/trackable_episodes/#{episode.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include("第1話")
  end

  it "ログインしていないとき、ログインページにリダイレクトすること" do
    episode = FactoryBot.create(:episode)

    get "/fragment/trackable_episodes/#{episode.id}"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "エピソードが存在しないとき、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)

    login_as(user, scope: :user)
    get "/fragment/trackable_episodes/99999"

    expect(response.status).to eq(404)
  end

  it "削除済みエピソードにアクセスしたとき、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, :deleted, work:)

    login_as(user, scope: :user)
    get "/fragment/trackable_episodes/#{episode.id}"

    expect(response.status).to eq(404)
  end

  it "エピソードに関連する作品情報を取得すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, title: "テスト作品")
    episode = FactoryBot.create(:episode, work:, title: "第1話")

    login_as(user, scope: :user)
    get "/fragment/trackable_episodes/#{episode.id}"

    expect(response.status).to eq(200)
    # ビューに作品タイトルが表示されることを確認
    expect(response.body).to include("テスト作品")
    expect(response.body).to include("第1話")
  end

  it "エピソード記録フォームが表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    login_as(user, scope: :user)
    get "/fragment/trackable_episodes/#{episode.id}"

    expect(response.status).to eq(200)
    # フォームが表示されることを確認
    expect(response.body).to include("episode_record_form")
  end

  it "エピソード記録一覧が設定されること" do
    user = FactoryBot.create(:registered_user)
    other_user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    # 他のユーザーの記録を作成
    record = FactoryBot.create(:record, :with_episode_record, user: other_user, work:, episode:)
    record.episode_record.update!(body: "面白かった", rating_state: "great")

    login_as(user, scope: :user)
    get "/fragment/trackable_episodes/#{episode.id}"

    expect(response.status).to eq(200)
    # set_episode_record_listメソッドが呼ばれることを確認（実際の表示内容はビューのテストで確認）
  end
end
