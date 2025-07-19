# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/channels/:channel_id/reception", type: :request do
  it "未ログイン時は401ステータスを返すこと" do
    channel = Channel.first
    post "/api/internal/channels/#{channel.id}/reception"

    expect(response.status).to eq(401)
  end

  it "ログイン時はチャンネルの受信を開始し201ステータスを返すこと" do
    user = create(:user)
    channel = Channel.first

    expect(user.receptions.exists?(channel:)).to be(false)

    login_as(user, scope: :user)
    post "/api/internal/channels/#{channel.id}/reception"

    expect(response.status).to eq(201)
    expect(user.receptions.exists?(channel:)).to be(true)
  end

  it "既に受信開始済みの場合でも201ステータスを返すこと" do
    user = create(:user)
    channel = Channel.first
    user.receive(channel)

    expect(user.receptions.exists?(channel:)).to be(true)

    login_as(user, scope: :user)
    post "/api/internal/channels/#{channel.id}/reception"

    expect(response.status).to eq(201)
    expect(user.receptions.exists?(channel:)).to be(true)
  end
end
