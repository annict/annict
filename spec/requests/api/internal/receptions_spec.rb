# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /api/internal/channels/:channel_id/reception", type: :request do
  it "未ログイン時は401ステータスを返すこと" do
    channel = Channel.first
    delete "/api/internal/channels/#{channel.id}/reception"

    expect(response.status).to eq(401)
  end

  it "ログイン時はチャンネルの受信を停止し200ステータスを返すこと" do
    user = create(:user)
    channel = Channel.first
    user.receive(channel)

    expect(user.receptions.exists?(channel:)).to be(true)

    login_as(user, scope: :user)
    delete "/api/internal/channels/#{channel.id}/reception"

    expect(response.status).to eq(200)
    expect(user.receptions.exists?(channel:)).to be(false)
  end

  it "既に受信停止済みの場合でも200ステータスを返すこと" do
    user = create(:user)
    channel = Channel.first

    expect(user.receptions.exists?(channel:)).to be(false)

    login_as(user, scope: :user)
    delete "/api/internal/channels/#{channel.id}/reception"

    expect(response.status).to eq(200)
    expect(user.receptions.exists?(channel:)).to be(false)
  end

  it "存在しないチャンネルIDの場合はActiveRecord::RecordNotFoundが発生すること" do
    user = create(:user)
    invalid_channel_id = SecureRandom.uuid

    login_as(user, scope: :user)

    expect do
      delete "/api/internal/channels/#{invalid_channel_id}/reception"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end
end
