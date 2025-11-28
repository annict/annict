# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/slots/:id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    _channel = Channel.create!(channel_group:, name: "テストチャンネル", sort_number: 1)
    slot = create(:slot, :published)

    get "/db/slots/#{slot.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディター権限がないユーザーでログインしているとき、アクセスできないこと" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    _channel = Channel.create!(channel_group:, name: "テストチャンネル", sort_number: 1)
    user = create(:registered_user)
    slot = create(:slot, :published)
    login_as(user, scope: :user)

    get "/db/slots/#{slot.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディター権限があるユーザーでログインしているとき、スロット編集フォームが表示されること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    _channel = Channel.create!(channel_group:, name: "テストチャンネル", sort_number: 1)
    user = create(:registered_user, :with_editor_role)
    slot = create(:slot, :published)
    login_as(user, scope: :user)

    get "/db/slots/#{slot.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(slot.channel.name)
  end
end
