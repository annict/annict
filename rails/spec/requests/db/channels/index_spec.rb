# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/channels", type: :request do
  it "ユーザーがログインしていないとき、チャンネル一覧を表示すること" do
    channel = create(:channel)

    get "/db/channels"

    expect(response.status).to eq(200)
    expect(response.body).to include(channel.name)
  end

  it "ユーザーがログインしているとき、チャンネル一覧を表示すること" do
    user = create(:registered_user)
    channel = create(:channel)
    login_as(user, scope: :user)

    get "/db/channels"

    expect(response.status).to eq(200)
    expect(response.body).to include(channel.name)
  end

  it "削除されたチャンネルは表示されないこと" do
    channel = create(:channel, deleted_at: Time.current)

    get "/db/channels"

    expect(response.status).to eq(200)
    expect(response.body).not_to include(channel.name)
  end

  it "VODチャンネルが先に表示されること" do
    non_vod_channel = create(:channel, sort_number: 1)
    vod_channel = create(:channel, :with_vod, sort_number: 2)

    get "/db/channels"

    expect(response.status).to eq(200)
    expect(response.body.index(vod_channel.name)).to be < response.body.index(non_vod_channel.name)
  end

  it "チャンネルグループと共に表示されること" do
    channel_group = create(:channel_group)
    channel = create(:channel, channel_group:)

    get "/db/channels"

    expect(response.status).to eq(200)
    expect(response.body).to include(channel.name)
    expect(response.body).to include(channel_group.name)
  end
end
