# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/channels", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel_params = {
      name: "ちゃんねる"
    }

    expect {
      post "/db/channels", params: {channel: channel_params}
    }.not_to change(Channel, :count)

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel_params = {
      name: "ちゃんねる"
    }

    login_as(user, scope: :user)

    expect {
      post "/db/channels", params: {channel: channel_params}
    }.not_to change(Channel, :count)

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限を持つユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel_params = {
      name: "ちゃんねる"
    }

    login_as(user, scope: :user)

    expect {
      post "/db/channels", params: {channel: channel_params}
    }.not_to change(Channel, :count)

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "管理者権限を持つユーザーがログインしているとき、チャンネルを作成できること" do
    channel_group = create(:channel_group)
    user = create(:registered_user, :with_admin_role)
    channel_params = {
      channel_group_id: channel_group.id,
      name: "ちゃんねる"
    }

    login_as(user, scope: :user)

    expect {
      post "/db/channels", params: {channel: channel_params}
    }.to change(Channel, :count).by(1)

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")

    channel = Channel.last

    expect(channel.channel_group_id).to eq(channel_group.id)
    expect(channel.name).to eq("ちゃんねる")
  end

  it "管理者権限を持つユーザーがログインしているとき、バリデーションエラーがある場合、作成に失敗すること" do
    user = create(:registered_user, :with_admin_role)
    channel_params = {
      channel_group_id: nil,
      name: ""
    }

    login_as(user, scope: :user)

    expect {
      post "/db/channels", params: {channel: channel_params}
    }.not_to change(Channel, :count)

    expect(response.status).to eq(422)
  end

  it "管理者権限を持つユーザーがログインしているとき、vodとsort_numberも設定できること" do
    channel_group = create(:channel_group)
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
