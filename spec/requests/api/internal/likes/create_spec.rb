# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/likes", type: :request do
  it "未ログイン時は401ステータスを返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    episode_record = FactoryBot.create(:episode_record, user:)

    post "/api/internal/likes", params: {
      recipient_type: "Record",
      recipient_id: episode_record.record.id
    }

    expect(response.status).to eq(401)
  end

  it "ログイン時はEpisodeRecordにいいねし201ステータスを返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    other_user = FactoryBot.create(:user, :with_email_notification)
    episode_record = FactoryBot.create(:episode_record, user: other_user)

    expect(user.likes.count).to eq(0)

    login_as(user, scope: :user)
    post "/api/internal/likes", params: {
      recipient_type: "Record",
      recipient_id: episode_record.record.id
    }

    expect(response.status).to eq(201)
    expect(user.likes.count).to eq(1)
    like = user.likes.first
    expect(like.recipient).to eq(episode_record)
  end

  it "ログイン時はWorkRecordにいいねし201ステータスを返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    other_user = FactoryBot.create(:user, :with_email_notification)
    work_record = FactoryBot.create(:work_record, user: other_user)

    expect(user.likes.count).to eq(0)

    login_as(user, scope: :user)
    post "/api/internal/likes", params: {
      recipient_type: "Record",
      recipient_id: work_record.record.id
    }

    expect(response.status).to eq(201)
    expect(user.likes.count).to eq(1)
    like = user.likes.first
    expect(like.recipient).to eq(work_record)
  end

  it "ログイン時はStatusにいいねし201ステータスを返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    other_user = FactoryBot.create(:user, :with_email_notification)
    status = FactoryBot.create(:status, user: other_user)

    expect(user.likes.count).to eq(0)

    login_as(user, scope: :user)
    post "/api/internal/likes", params: {
      recipient_type: "Status",
      recipient_id: status.id
    }

    expect(response.status).to eq(201)
    expect(user.likes.count).to eq(1)
    like = user.likes.first
    expect(like.recipient).to eq(status)
  end

  it "既にいいね済みの場合でも201ステータスを返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    other_user = FactoryBot.create(:user, :with_email_notification)
    episode_record = FactoryBot.create(:episode_record, user: other_user)
    FactoryBot.create(:like, user:, recipient: episode_record)

    expect(user.likes.count).to eq(1)

    login_as(user, scope: :user)
    post "/api/internal/likes", params: {
      recipient_type: "Record",
      recipient_id: episode_record.record.id
    }

    expect(response.status).to eq(201)
    expect(user.likes.count).to eq(1)
  end

  it "存在しないrecipient_idを指定した場合は404エラーが返されること" do
    user = FactoryBot.create(:user, :with_email_notification)

    login_as(user, scope: :user)
    post "/api/internal/likes", params: {
      recipient_type: "Record",
      recipient_id: "nonexistent"
    }

    expect(response.status).to eq(404)
  end

  it "不正なrecipient_typeを指定した場合はNameErrorが発生すること" do
    user = FactoryBot.create(:user, :with_email_notification)

    login_as(user, scope: :user)
    expect {
      post "/api/internal/likes", params: {
        recipient_type: "InvalidType",
        recipient_id: "123"
      }
    }.to raise_error(NameError)
  end

  it "recipient_typeが不足している場合はNameErrorが発生すること" do
    user = FactoryBot.create(:user, :with_email_notification)

    login_as(user, scope: :user)
    expect {
      post "/api/internal/likes", params: {
        recipient_id: "123"
      }
    }.to raise_error(NameError)
  end

  it "recipient_idが不足している場合は404エラーが返されること" do
    user = FactoryBot.create(:user, :with_email_notification)

    login_as(user, scope: :user)
    post "/api/internal/likes", params: {
      recipient_type: "Record"
    }

    expect(response.status).to eq(404)
  end

  it "自分のコンテンツにいいねした場合でも201ステータスを返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    episode_record = FactoryBot.create(:episode_record, user:)

    expect(user.likes.count).to eq(0)

    login_as(user, scope: :user)
    post "/api/internal/likes", params: {
      recipient_type: "Record",
      recipient_id: episode_record.record.id
    }

    expect(response.status).to eq(201)
    expect(user.likes.count).to eq(1)
    like = user.likes.first
    expect(like.recipient).to eq(episode_record)
  end
end
