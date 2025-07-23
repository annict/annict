# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/channels/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel = Channel.first
    old_channel = channel.attributes
    channel_params = {
      name: "ちゃんねる"
    }

    patch "/db/channels/#{channel.id}", params: {channel: channel_params}
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(channel.name).to eq(old_channel["name"])
  end

  it "編集者権限のユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel = Channel.first
    old_channel = channel.attributes
    channel_params = {
      name: "ちゃんねる"
    }

    login_as(user, scope: :user)

    patch "/db/channels/#{channel.id}", params: {channel: channel_params}
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(channel.name).to eq(old_channel["name"])
  end

  it "一般ユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel = Channel.first
    old_channel = channel.attributes
    channel_params = {
      name: "ちゃんねる"
    }

    login_as(user, scope: :user)

    patch "/db/channels/#{channel.id}", params: {channel: channel_params}
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(channel.name).to eq(old_channel["name"])
  end

  it "管理者がログインしているとき、チャンネルを更新できること" do
    user = create(:registered_user, :with_admin_role)
    channel = Channel.first
    old_channel = channel.attributes
    channel_params = {
      name: "ちゃんねる"
    }

    login_as(user, scope: :user)

    expect(channel.name).to eq(old_channel["name"])

    patch "/db/channels/#{channel.id}", params: {channel: channel_params}
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(channel.name).to eq("ちゃんねる")
  end

  it "管理者がログインしているとき、全てのパラメータを更新できること" do
    user = create(:registered_user, :with_admin_role)
    channel = Channel.first
    channel_group = ChannelGroup.second
    channel_params = {
      name: "新しいチャンネル名",
      channel_group_id: channel_group.id,
      vod: true,
      sort_number: 100
    }

    login_as(user, scope: :user)

    patch "/db/channels/#{channel.id}", params: {channel: channel_params}
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(channel.name).to eq("新しいチャンネル名")
    expect(channel.channel_group_id).to eq(channel_group.id)
    expect(channel.vod).to eq(true)
    expect(channel.sort_number).to eq(100)
  end

  it "管理者がログインしているとき、必須パラメータが不正な場合、更新に失敗すること" do
    user = create(:registered_user, :with_admin_role)
    channel = Channel.first
    old_channel = channel.attributes
    channel_params = {
      name: "",
      channel_group_id: channel.channel_group_id
    }

    login_as(user, scope: :user)

    patch "/db/channels/#{channel.id}", params: {channel: channel_params}
    channel.reload

    expect(response.status).to eq(422)
    expect(channel.name).to eq(old_channel["name"])
  end

  it "管理者がログインしているとき、存在しないチャンネルIDの場合、404エラーになること" do
    user = create(:registered_user, :with_admin_role)
    channel_params = {
      name: "ちゃんねる"
    }

    login_as(user, scope: :user)

    expect {
      patch "/db/channels/non-existent-id", params: {channel: channel_params}
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "管理者がログインしているとき、削除されたチャンネルは更新できないこと" do
    user = create(:registered_user, :with_admin_role)
    channel = Channel.first
    channel.update!(deleted_at: Time.current)
    channel_params = {
      name: "ちゃんねる"
    }

    login_as(user, scope: :user)

    expect {
      patch "/db/channels/#{channel.id}", params: {channel: channel_params}
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
