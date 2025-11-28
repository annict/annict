# typed: false
# frozen_string_literal: true

RSpec.describe "GET /api/internal/received_channels", type: :request do
  it "未ログイン時は空の配列を返すこと" do
    get "/api/internal/received_channels"

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq([])
  end

  it "ログイン時でチャンネルを受信していない場合は空の配列を返すこと" do
    user = create(:user, :with_email_notification)

    login_as(user, scope: :user)
    get "/api/internal/received_channels"

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq([])
  end

  it "ログイン時でチャンネルを受信している場合はチャンネルIDの配列を返すこと" do
    user = create(:user, :with_email_notification)
    channel_group = ChannelGroup.create!(name: "テストチャンネルグループ")
    channel1 = Channel.create!(channel_group:, name: "チャンネル1")
    channel2 = Channel.create!(channel_group:, name: "チャンネル2")
    Reception.create!(user:, channel: channel1)
    Reception.create!(user:, channel: channel2)

    login_as(user, scope: :user)
    get "/api/internal/received_channels"

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body).to match_array([channel1.id, channel2.id])
  end

  it "複数のチャンネルを受信している場合は全てのチャンネルIDが含まれること" do
    user = create(:user, :with_email_notification)
    channel_group = ChannelGroup.create!(name: "テストチャンネルグループ")
    channels = 5.times.map do |i|
      Channel.create!(channel_group:, name: "チャンネル#{i + 1}")
    end
    channels.each do |channel|
      Reception.create!(user:, channel:)
    end

    login_as(user, scope: :user)
    get "/api/internal/received_channels"

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body).to match_array(channels.map(&:id))
  end

  it "他のユーザーの受信チャンネルは含まれないこと" do
    user1 = create(:user, :with_email_notification)
    user2 = create(:user, :with_email_notification)
    channel_group = ChannelGroup.create!(name: "テストチャンネルグループ")
    channel1 = Channel.create!(channel_group:, name: "チャンネル1")
    channel2 = Channel.create!(channel_group:, name: "チャンネル2")
    Reception.create!(user: user1, channel: channel1)
    Reception.create!(user: user2, channel: channel2)

    login_as(user1, scope: :user)
    get "/api/internal/received_channels"

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body).to eq([channel1.id])
  end
end
