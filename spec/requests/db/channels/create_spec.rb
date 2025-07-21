# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/channels", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel_params = {
      name: "ちゃんねる"
    }

    post "/db/channels", params: {channel: channel_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")

    expect(Channel.all.size).to eq(220)
  end

  it "エディターロールを持たないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel_params = {
      name: "ちゃんねる"
    }

    login_as(user, scope: :user)

    post "/db/channels", params: {channel: channel_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(Channel.all.size).to eq(220)
  end

  it "エディターロールを持つユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel_params = {
      name: "ちゃんねる"
    }

    login_as(user, scope: :user)

    post "/db/channels", params: {channel: channel_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(Channel.all.size).to eq(220)
  end

  it "管理者ロールを持つユーザーがログインしているとき、チャンネルを作成できること" do
    channel_group = ChannelGroup.first
    user = create(:registered_user, :with_admin_role)
    channel_params = {
      channel_group_id: channel_group.id,
      name: "ちゃんねる"
    }

    login_as(user, scope: :user)

    expect(Channel.all.size).to eq(220)

    post "/db/channels", params: {channel: channel_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")

    expect(Channel.all.size).to eq(221)
    channel = Channel.last

    expect(channel.channel_group_id).to eq(channel_group.id)
    expect(channel.name).to eq("ちゃんねる")
  end

  it "管理者ロールを持つユーザーがログインしているとき、バリデーションエラーがある場合、作成に失敗すること" do
    user = create(:registered_user, :with_admin_role)
    channel_params = {
      channel_group_id: nil,
      name: ""
    }

    login_as(user, scope: :user)

    expect(Channel.all.size).to eq(220)

    post "/db/channels", params: {channel: channel_params}

    expect(response.status).to eq(422)
    expect(Channel.all.size).to eq(220)
  end

  it "管理者ロールを持つユーザーがログインしているとき、vodとsort_numberも設定できること" do
    channel_group = ChannelGroup.first
    user = create(:registered_user, :with_admin_role)
    channel_params = {
      channel_group_id: channel_group.id,
      name: "ちゃんねる",
      vod: true,
      sort_number: 100
    }

    login_as(user, scope: :user)

    post "/db/channels", params: {channel: channel_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")

    channel = Channel.last
    expect(channel.channel_group_id).to eq(channel_group.id)
    expect(channel.name).to eq("ちゃんねる")
    expect(channel.vod).to eq(true)
    expect(channel.sort_number).to eq(100)
  end
end
