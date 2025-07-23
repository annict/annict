# typed: false
# frozen_string_literal: true

RSpec.describe "GET /notifications", type: :request do
  it "ログインユーザーが通知一覧を表示できること" do
    user_1 = FactoryBot.create(:registered_user)
    user_2 = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user: user_1, work:)
    FactoryBot.create(:work_record, user: user_1, work:, record:)

    Creators::LikeCreator.new(user: user_2, likeable: record).call

    login_as(user_1, scope: :user)
    get "/notifications"

    expect(response.status).to eq(200)
    expect(response.body).to include(user_2.profile.name)
    expect(response.body).to include("に「いいね！」しました")
  end

  it "未読通知がある場合、通知が既読になること" do
    user_1 = FactoryBot.create(:registered_user)
    user_2 = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user: user_1, work:)
    FactoryBot.create(:work_record, user: user_1, work:, record:)

    # 通知を作成（いいね！することで通知が生成される）
    Creators::LikeCreator.new(user: user_2, likeable: record).call

    # 通知の存在と未読状態を確認
    user_1.reload
    notification = user_1.notifications.first
    expect(notification.read).to eq(false)
    expect(user_1.notifications_count).to be > 0

    login_as(user_1, scope: :user)
    get "/notifications"

    expect(response.status).to eq(200)
    expect(notification.reload.read).to eq(true)
  end

  it "通知がない場合、空の状態が表示されること" do
    user = FactoryBot.create(:registered_user)

    login_as(user, scope: :user)
    get "/notifications"

    expect(response.status).to eq(200)
    expect(response.body).to include("通知はありません")
  end

  it "ログインしていない場合、ログインページにリダイレクトされること" do
    get "/notifications"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "ページネーションが機能すること" do
    user_1 = FactoryBot.create(:registered_user)

    # 31件の通知を作成（ページ2が存在するように）
    31.times do |i|
      user_2 = FactoryBot.create(:registered_user)
      work = FactoryBot.create(:work)
      record = FactoryBot.create(:record, user: user_1, work:)
      FactoryBot.create(:work_record, user: user_1, work:, record:)
      Creators::LikeCreator.new(user: user_2, likeable: record).call
    end

    login_as(user_1, scope: :user)
    get "/notifications", params: {page: 2}

    expect(response.status).to eq(200)
  end
end
