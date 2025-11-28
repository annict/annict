# typed: false
# frozen_string_literal: true

RSpec.describe "GET /api/internal/likes", type: :request do
  it "未ログイン時は空の配列を返すこと" do
    get "/api/internal/likes"

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq([])
  end

  it "ログイン時は自分のいいねのリストを返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    other_user = FactoryBot.create(:user, :with_email_notification)

    # EpisodeRecordのいいねを作成
    # EpisodeRecordのrecipient_typeはRecordとして返される
    episode_record = FactoryBot.create(:episode_record, user: other_user)
    FactoryBot.create(:like, user:, recipient: episode_record)

    # WorkRecordのいいねを作成
    # WorkRecordのrecipient_typeはRecordとして返される
    work_record = FactoryBot.create(:work_record, user: other_user)
    FactoryBot.create(:like, user:, recipient: work_record)

    # Statusのいいねを作成
    status = FactoryBot.create(:status, user: other_user)
    FactoryBot.create(:like, user:, recipient: status)

    # 他のユーザーのいいねを作成（取得されないはず）
    other_episode_record = FactoryBot.create(:episode_record, user:)
    FactoryBot.create(:like, user: other_user, recipient: other_episode_record)

    login_as(user, scope: :user)
    get "/api/internal/likes"

    expect(response.status).to eq(200)

    likes = JSON.parse(response.body)
    expect(likes.size).to eq(3)

    # レスポンスの内容を検証
    # EpisodeRecordとWorkRecordはRecordとして返される
    expected_likes = [
      {
        "recipient_type" => "Record",
        "recipient_id" => episode_record.record.id
      },
      {
        "recipient_type" => "Record",
        "recipient_id" => work_record.record.id
      },
      {
        "recipient_type" => "Status",
        "recipient_id" => status.id
      }
    ]

    expect(likes).to match_array(expected_likes)
  end

  it "WorkRecord（AnimeRecord）の場合もRecordを返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    other_user = FactoryBot.create(:user, :with_email_notification)

    # WorkRecord（AnimeRecordはWorkRecordのエイリアス）はRecordとして扱われる
    work_record = FactoryBot.create(:work_record, user: other_user)
    FactoryBot.create(:like, user:, recipient: work_record)

    login_as(user, scope: :user)
    get "/api/internal/likes"

    expect(response.status).to eq(200)

    likes = JSON.parse(response.body)
    expect(likes.size).to eq(1)
    expect(likes.first).to eq({
      "recipient_type" => "Record",
      "recipient_id" => work_record.record.id
    })
  end

  it "いいねがない場合は空の配列を返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)

    login_as(user, scope: :user)
    get "/api/internal/likes"

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq([])
  end
end
