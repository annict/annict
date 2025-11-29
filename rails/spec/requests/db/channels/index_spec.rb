# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/channels", type: :request do
  it "ユーザーがログインしていないとき、チャンネル一覧を表示すること" do
    channel = Channel.first

    get "/db/channels"

    expect(response.status).to eq(200)
    expect(response.body).to include(channel.name)
  end

  it "ユーザーがログインしているとき、チャンネル一覧を表示すること" do
    user = create(:registered_user)
    channel = Channel.first
    login_as(user, scope: :user)

    get "/db/channels"

    expect(response.status).to eq(200)
    expect(response.body).to include(channel.name)
  end

  it "削除されたチャンネルは表示されないこと" do
    channel = Channel.first
    channel.update!(deleted_at: Time.current)

    get "/db/channels"

    expect(response.status).to eq(200)
    expect(response.body).not_to include(channel.name)
  end

  it "VODチャンネルが先に表示されること" do
    # VODでないチャンネルとVODチャンネルが存在することを確認
    vod_channel = Channel.find_by(vod: true)
    non_vod_channel = Channel.find_by(vod: false)

    if vod_channel && non_vod_channel
      get "/db/channels"

      expect(response.status).to eq(200)
      # VODチャンネルが先に表示されることを確認
      vod_position = response.body.index(vod_channel.name)
      non_vod_position = response.body.index(non_vod_channel.name)
      expect(vod_position).to be < non_vod_position if vod_position && non_vod_position
    end
  end

  it "チャンネルグループと共に表示されること" do
    channel = Channel.eager_load(:channel_group).merge(ChannelGroup.without_deleted).first

    if channel&.channel_group
      get "/db/channels"

      expect(response.status).to eq(200)
      expect(response.body).to include(channel.name)
    end
  end
end
