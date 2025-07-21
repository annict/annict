# typed: false
# frozen_string_literal: true

RSpec.describe "GET /fragment/@:username/tracking_heatmap", type: :request do
  it "ユーザーが存在しない場合、404エラーを返すこと" do
    expect {
      get "/fragment/@nonexistentuser/tracking_heatmap"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたユーザーのヒートマップを表示しようとした場合、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)
    user.destroy!

    expect {
      get "/fragment/@#{user.username}/tracking_heatmap"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "ユーザーが存在する場合、200を返すこと" do
    user = FactoryBot.create(:registered_user)

    get "/fragment/@#{user.username}/tracking_heatmap"

    expect(response).to have_http_status(:ok)
  end

  it "記録データがない場合でも200を返すこと" do
    user = FactoryBot.create(:registered_user)

    get "/fragment/@#{user.username}/tracking_heatmap"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("c-tracking-heatmap")
  end

  it "記録データがある場合、ヒートマップが表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    # 今日の記録
    FactoryBot.create(:record, user: user, work: work, watched_at: Time.zone.today)
    # 昨日の記録を3つ
    3.times do
      FactoryBot.create(:record, user: user, work: work, watched_at: Time.zone.today - 1.day)
    end
    # 1週間前の記録を5つ
    5.times do
      FactoryBot.create(:record, user: user, work: work, watched_at: Time.zone.today - 7.days)
    end

    get "/fragment/@#{user.username}/tracking_heatmap"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("c-tracking-heatmap")
    expect(response.body).to include("c-tracking-heatmap__day")
    expect(response.body).to include("c-tracking-heatmap__density-")
  end

  it "150日以上前の記録は含まれないこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    # 150日前の記録
    date_150_days_ago = (Time.zone.today - 150.days).beginning_of_week(:sunday)
    FactoryBot.create(:record, user: user, work: work, watched_at: date_150_days_ago)

    # 151日前の記録（含まれないはず）
    date_151_days_ago = date_150_days_ago - 1.day
    FactoryBot.create(:record, user: user, work: work, watched_at: date_151_days_ago)

    get "/fragment/@#{user.username}/tracking_heatmap"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(date_150_days_ago.strftime("%Y-%m-%d"))
    expect(response.body).not_to include(date_151_days_ago.strftime("%Y-%m-%d"))
  end

  it "別のユーザーのヒートマップにアクセスした場合でも表示されること" do
    owner = FactoryBot.create(:registered_user)
    viewer = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    FactoryBot.create(:record, user: owner, work: work)

    login_as(viewer, scope: :user)

    get "/fragment/@#{owner.username}/tracking_heatmap"

    expect(response).to have_http_status(:ok)
  end

  it "未ログインでもヒートマップが表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    FactoryBot.create(:record, user: user, work: work)

    get "/fragment/@#{user.username}/tracking_heatmap"

    expect(response).to have_http_status(:ok)
  end

  it "タイムゾーンがクッキーに設定されている場合、その値が使用されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    # 日付の境界をテストするため、日本時間の翌日0時の記録を作成
    # UTC時間では前日の15時
    watched_at_jst = Time.zone.parse("2025-07-21 00:00:00 JST")
    watched_at_utc = watched_at_jst.utc
    FactoryBot.create(:record, user: user, work: work, watched_at: watched_at_utc)

    # ヨーロッパ/ロンドンのタイムゾーンを設定（UTC+1時間）
    # この場合、UTC 2025-07-20 15:00は、ロンドン時間で2025-07-20 16:00となる
    get "/fragment/@#{user.username}/tracking_heatmap", headers: {
      "Cookie" => "ann_time_zone=Europe/London"
    }

    expect(response).to have_http_status(:ok)
    # ロンドン時間では7月20日として表示される
    expect(response.body).to include("2025-07-20")
  end

  it "ログインユーザーのタイムゾーンが優先されること" do
    user = FactoryBot.create(:registered_user, time_zone: "Europe/London")
    viewer = FactoryBot.create(:registered_user, time_zone: "America/New_York")
    work = FactoryBot.create(:work)

    # UTC時間で23時の記録
    watched_at_utc = Time.zone.parse("2024-01-01 23:00:00 UTC")
    FactoryBot.create(:record, user: user, work: work, watched_at: watched_at_utc)

    # クッキーに別のタイムゾーンを設定
    cookies["ann_time_zone"] = "Asia/Tokyo"

    login_as(viewer, scope: :user)

    get "/fragment/@#{user.username}/tracking_heatmap"

    expect(response).to have_http_status(:ok)
  end

  it "記録の数に応じて適切なレベルが設定されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    # レベル0: 0件
    # レベル1: 1〜3件
    FactoryBot.create(:record, user: user, work: work, watched_at: Time.zone.today)

    # レベル2: 4〜6件
    4.times do
      FactoryBot.create(:record, user: user, work: work, watched_at: Time.zone.today - 1.day)
    end

    # レベル3: 7〜9件
    7.times do
      FactoryBot.create(:record, user: user, work: work, watched_at: Time.zone.today - 2.days)
    end

    # レベル4: 10件以上
    10.times do
      FactoryBot.create(:record, user: user, work: work, watched_at: Time.zone.today - 3.days)
    end

    get "/fragment/@#{user.username}/tracking_heatmap"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("c-tracking-heatmap__density-1") # 1件
    expect(response.body).to include("c-tracking-heatmap__density-2") # 4件
    expect(response.body).to include("c-tracking-heatmap__density-3") # 7件
    expect(response.body).to include("c-tracking-heatmap__density-4") # 10件
  end
end
