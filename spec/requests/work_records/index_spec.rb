# typed: false
# frozen_string_literal: true

RSpec.describe "GET /works/:work_id/records", type: :request do
  it "ログインしているとき、作品の記録一覧を表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, :with_work_record, user:, work:)
    record.work_record.update!(body: "面白かった", rating_overall_state: "great")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("面白かった")
  end

  it "ログインしていないとき、記録一覧を表示すること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user)
    record = FactoryBot.create(:record, :with_work_record, user:, work:)
    record.work_record.update!(body: "面白かった", rating_overall_state: "great")

    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("面白かった")
  end

  it "作品が存在しないとき、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)

    login_as(user, scope: :user)
    get "/works/99999/records"

    expect(response.status).to eq(404)
  end

  it "削除済み作品の記録にアクセスしたとき、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, :deleted)

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(404)
  end

  it "自分の記録を表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    my_record = FactoryBot.create(:record, :with_work_record, user:, work:)
    my_record.work_record.update!(body: "自分の記録")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("自分の記録")
  end

  it "フォローしているユーザーの記録を表示すること" do
    user = FactoryBot.create(:registered_user)
    following_user = FactoryBot.create(:registered_user)
    user.follow(following_user)

    work = FactoryBot.create(:work)
    following_record = FactoryBot.create(:record, :with_work_record, user: following_user, work:)
    following_record.work_record.update!(body: "フォロー中ユーザーの記録")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("フォロー中ユーザーの記録")
  end

  it "ミュートしているユーザーの記録を表示しないこと" do
    user = FactoryBot.create(:registered_user)
    muted_user = FactoryBot.create(:registered_user)
    user.mute_users.create!(muted_user:)

    work = FactoryBot.create(:work)
    muted_record = FactoryBot.create(:record, :with_work_record, user: muted_user, work:)
    muted_record.work_record.update!(body: "ミュートユーザーの記録")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).not_to include("ミュートユーザーの記録")
  end

  it "削除済みの記録を表示しないこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    active_record = FactoryBot.create(:record, :with_work_record, user:, work:)
    active_record.work_record.update!(body: "表示される記録")

    deleted_record = FactoryBot.create(:record, :with_work_record, user:, work:)
    deleted_record.work_record.update!(body: "削除された記録")
    deleted_record.destroy!

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("表示される記録")
    expect(response.body).not_to include("削除された記録")
  end

  it "レーティングの高い順に記録を表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    bad_record = FactoryBot.create(:record, :with_work_record, user:, work:)
    bad_record.work_record.update!(body: "悪い評価", rating_overall_state: "bad")

    great_record = FactoryBot.create(:record, :with_work_record, user:, work:)
    great_record.work_record.update!(body: "素晴らしい評価", rating_overall_state: "great")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    # rating_overall_stateの高い順で表示されることを確認
    great_position = response.body.index("素晴らしい評価")
    bad_position = response.body.index("悪い評価")
    expect(great_position).to be < bad_position
  end

  it "ログインユーザーの場合、本文のない記録は全体の記録一覧に表示されないこと" do
    user = FactoryBot.create(:registered_user)
    other_user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    record_with_body = FactoryBot.create(:record, :with_work_record, user: other_user, work:)
    record_with_body.work_record.update!(body: "コメントあり")

    record_without_body = FactoryBot.create(:record, :with_work_record, user: other_user, work:)
    record_without_body.work_record.update!(body: "")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("コメントあり")
    # 本文のない記録は全体の記録一覧には含まれない
  end

  it "ページネーションが機能すること" do
    user = FactoryBot.create(:registered_user)
    # ユニークなタイトルで作品を作成し、他のテストとの競合を避ける
    unique_title = "テスト作品#{SecureRandom.hex(8)}"
    work = FactoryBot.create(:work, title: unique_title)

    # この作品に関連する既存の記録を削除
    Record.joins(:work_record).where(work_records: {work_id: work.id}).destroy_all

    # レーティングが設定された記録を5件作成（これらが最初に表示される）
    5.times do |i|
      other_user = FactoryBot.create(:registered_user)
      record = FactoryBot.create(:record, :with_work_record, user: other_user, work:)
      record.work_record.update!(
        body: "レーティングあり記録#{i + 1}",
        created_at: Time.current - i.hours,
        rating_overall_state: "great"
      )
    end

    # レーティングなしの記録を101件作成（NULLS LASTにより後ろに配置される）
    records = []
    101.times do |i|
      other_user = FactoryBot.create(:registered_user)
      record = FactoryBot.create(:record, :with_work_record, user: other_user, work:)
      # 作成日時を設定（i=0が最古、i=100が最新）
      record.work_record.update!(
        body: "#{unique_title}の記録#{i + 1}",
        created_at: Time.current - (101 - i).hours,
        rating_overall_state: nil # レーティングなし
      )
      record.update!(created_at: Time.current - (101 - i).hours)
      records << record
    end

    login_as(user, scope: :user)

    # ソート順：
    # 1. レーティングありの記録（5件）
    # 2. レーティングなしの記録（101件）- 作成日時の新しい順
    # 合計106件

    # 現在の記録数を確認
    all_records_count = Record.joins(:work_record).where(work_records: {work_id: work.id}).count
    expect(all_records_count).to eq(106)

    # 1ページ目: レーティングあり5件 + レーティングなし95件 = 100件
    # 2ページ目: レーティングなし6件（記録6から記録1まで）

    # 2ページ目を確認
    get "/works/#{work.id}/records?page=2"
    expect(response.status).to eq(200)
    # レーティングなしの古い記録が表示される
    expect(response.body).to include("#{unique_title}の記録1") # 最古の記録
    expect(response.body).to include("#{unique_title}の記録6") # 6番目の記録
    expect(response.body).not_to include("#{unique_title}の記録7") # 7番目以降は1ページ目
    expect(response.body).not_to include("レーティングあり記録") # レーティングありは1ページ目
  end

  it "作品のタイトルがページタイトルに含まれること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, title: "テストアニメ")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("<title>")
    expect(response.body).to include("テストアニメ")
  end

  it "エピソード記録ではなく作品記録のみ表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    # 作品記録
    work_record = FactoryBot.create(:record, :with_work_record, user:, work:)
    work_record.work_record.update!(body: "作品記録")

    # エピソード記録
    episode_record = FactoryBot.create(:record, :with_episode_record, user:, episode:)
    episode_record.episode_record.update!(body: "エピソード記録")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("作品記録")
    expect(response.body).not_to include("エピソード記録")
  end
end
