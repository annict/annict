# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/channel_groups", type: :request do
  it "ログインしていない場合、アクセスできないこと" do
    channel_group_params = {
      name: "ちゃんねるぐるーぷ"
    }

    post "/db/channel_groups", params: {channel_group: channel_group_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")

    expect(ChannelGroup.all.size).to eq(18)
  end

  it "エディター権限を持たないユーザーがログインしている場合、アクセスできないこと" do
    user = create(:registered_user)
    channel_group_params = {
      name: "ちゃんねるぐるーぷ"
    }

    login_as(user, scope: :user)

    post "/db/channel_groups", params: {channel_group: channel_group_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(ChannelGroup.all.size).to eq(18)
  end

  it "エディター権限を持つユーザーがログインしている場合、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel_group_params = {
      name: "ちゃんねるぐるーぷ"
    }

    login_as(user, scope: :user)

    post "/db/channel_groups", params: {channel_group: channel_group_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(ChannelGroup.all.size).to eq(18)
  end

  it "管理者権限を持つユーザーがログインしている場合、チャンネルグループを作成できること" do
    user = create(:registered_user, :with_admin_role)
    channel_group_params = {
      name: "ちゃんねるぐるーぷ",
      sort_number: 10
    }

    login_as(user, scope: :user)

    expect(ChannelGroup.all.size).to eq(18)

    post "/db/channel_groups", params: {channel_group: channel_group_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")

    expect(ChannelGroup.all.size).to eq(19)
    channel_group = ChannelGroup.last

    expect(channel_group.name).to eq("ちゃんねるぐるーぷ")
    expect(channel_group.sort_number).to eq(10)
  end
end
