# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/channel_groups", type: :request do
  it "ログインしていないとき、チャンネルグループ一覧を表示すること" do
    channel_group = ChannelGroup.first

    get "/db/channel_groups"

    expect(response.status).to eq(200)
    expect(response.body).to include(channel_group.name)
  end

  it "ログインしているとき、チャンネルグループ一覧を表示すること" do
    user = create(:registered_user)
    channel_group = ChannelGroup.first
    login_as(user, scope: :user)

    get "/db/channel_groups"

    expect(response.status).to eq(200)
    expect(response.body).to include(channel_group.name)
  end
end
