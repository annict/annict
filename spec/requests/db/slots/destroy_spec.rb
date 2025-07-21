# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/slots/:id", type: :request do
  it "ログインしていないユーザーは削除できないこと" do
    channel_group = ChannelGroup.create!(name: "テストチャンネルグループ")
    channel = Channel.create!(channel_group:, name: "テストチャンネル")
    work = create(:work)
    episode = create(:episode, work:)
    program = create(:program, work:, channel:)
    slot = create(:slot, :not_deleted, work:, episode:, program:, channel:)

    expect(Slot.count).to eq(1)

    delete "/db/slots/#{slot.id}"
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")

    expect(Slot.count).to eq(1)
  end

  it "一般ユーザーはスロットを削除できないこと" do
    user = create(:registered_user)
    channel_group = ChannelGroup.create!(name: "テストチャンネルグループ")
    channel = Channel.create!(channel_group:, name: "テストチャンネル")
    work = create(:work)
    episode = create(:episode, work:)
    program = create(:program, work:, channel:)
    slot = create(:slot, :not_deleted, work:, episode:, program:, channel:)
    login_as(user, scope: :user)

    expect(Slot.count).to eq(1)

    delete "/db/slots/#{slot.id}"
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(Slot.count).to eq(1)
  end

  it "エディター権限ユーザーはスロットを削除できないこと" do
    user = create(:registered_user, :with_editor_role)
    channel_group = ChannelGroup.create!(name: "テストチャンネルグループ")
    channel = Channel.create!(channel_group:, name: "テストチャンネル")
    work = create(:work)
    episode = create(:episode, work:)
    program = create(:program, work:, channel:)
    slot = create(:slot, :not_deleted, work:, episode:, program:, channel:)
    login_as(user, scope: :user)

    expect(Slot.count).to eq(1)

    delete "/db/slots/#{slot.id}"
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(Slot.count).to eq(1)
  end

  it "管理者はスロットをソフト削除できること" do
    user = create(:registered_user, :with_admin_role)
    channel_group = ChannelGroup.create!(name: "テストチャンネルグループ")
    channel = Channel.create!(channel_group:, name: "テストチャンネル")
    work = create(:work)
    episode = create(:episode, work:)
    program = create(:program, work:, channel:)
    slot = create(:slot, :not_deleted, work:, episode:, program:, channel:)
    login_as(user, scope: :user)

    expect(Slot.count).to eq(1)

    delete "/db/slots/#{slot.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")

    expect(Slot.count).to eq(0)
  end
end
