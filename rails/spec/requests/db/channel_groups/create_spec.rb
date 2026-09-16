# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/channel_groups", type: :request do
  it "ログインしていない場合、アクセスできないこと" do
    channel_group_params = {
      name: "ちゃんねるぐるーぷ"
    }

    expect {
      post "/db/channel_groups", params: {channel_group: channel_group_params}
    }.not_to change(ChannelGroup, :count)

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限を持たないユーザーがログインしている場合、アクセスできないこと" do
    user = create(:registered_user)
    channel_group_params = {
      name: "ちゃんねるぐるーぷ"
    }

    login_as(user, scope: :user)

    expect {
      post "/db/channel_groups", params: {channel_group: channel_group_params}
    }.not_to change(ChannelGroup, :count)

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限を持つユーザーがログインしている場合、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel_group_params = {
      name: "ちゃんねるぐるーぷ"
    }

    login_as(user, scope: :user)

    expect {
      post "/db/channel_groups", params: {channel_group: channel_group_params}
    }.not_to change(ChannelGroup, :count)

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "管理者権限を持つユーザーがログインしている場合、チャンネルグループを作成できること" do
    user = create(:registered_user, :with_admin_role)
    channel_group_params = {
      name: "ちゃんねるぐるーぷ",
      sort_number: 10
    }

    login_as(user, scope: :user)

    expect {
      post "/db/channel_groups", params: {channel_group: channel_group_params}
    }.to change(ChannelGroup, :count).by(1)

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")

    channel_group = ChannelGroup.last

    expect(channel_group.name).to eq("ちゃんねるぐるーぷ")
    expect(channel_group.sort_number).to eq(10)
  end
end
