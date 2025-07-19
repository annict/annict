# typed: false
# frozen_string_literal: true

RSpec.describe "GET /@:username/records", type: :request do
  it "ユーザーがログインしていて記録が存在しない場合、記録がないメッセージを表示すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/@#{user.username}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("記録はありません")
  end

  it "ユーザーがログインしていて記録が存在する場合、記録を表示すること" do
    user = create(:registered_user)
    record_1 = create(:record, user:)
    create(:episode_record, record: record_1, user:, body: "楽しかった")
    record_2 = create(:record, user:)
    create(:work_record, user:, record: record_2, body: "最高")
    login_as(user, scope: :user)

    get "/@#{user.username}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include(user.profile.name)
    expect(response.body).to include("楽しかった")
    expect(response.body).to include("最高")
  end

  it "ユーザーがログインしていない場合で記録が存在しない場合、記録がないメッセージを表示すること" do
    user = create(:registered_user)

    get "/@#{user.username}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("記録はありません")
  end

  it "ユーザーがログインしていない場合で記録が存在する場合、記録を表示すること" do
    user = create(:registered_user)
    record_1 = create(:record, user:)
    create(:episode_record, record: record_1, user:, body: "楽しかった")
    record_2 = create(:record, user:)
    create(:work_record, user:, record: record_2, body: "最高")

    get "/@#{user.username}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include(user.profile.name)
    expect(response.body).to include("楽しかった")
    expect(response.body).to include("最高")
  end

  it "月パラメーターが指定された場合、指定された月の記録のみを表示すること" do
    user = create(:registered_user)
    record_1 = create(:record, user:, watched_at: Time.zone.parse("2020-04-01"))
    create(:work_record, user:, record: record_1, body: "最高")
    record_2 = create(:record, user:, watched_at: Time.zone.parse("2020-05-01"))
    create(:work_record, user:, record: record_2, body: "すごく良かった")

    get "/@#{user.username}/records?year=2020&month=5"

    expect(response.status).to eq(200)
    expect(response.body).to include(user.profile.name)
    expect(response.body).to include("すごく良かった")
    expect(response.body).not_to include("最高")
  end
end
