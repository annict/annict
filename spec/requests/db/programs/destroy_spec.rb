# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/programs/:id", type: :request do
  it "ユーザーがログインしていないとき、ログインページにリダイレクトすること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = create(:program, :not_deleted, channel:)

    expect(Program.count).to eq(1)

    delete "/db/programs/#{program.id}"
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")

    expect(Program.count).to eq(1)
  end

  it "通常ユーザーがログインしているとき、アクセス拒否されること" do
    user = create(:registered_user)
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = create(:program, :not_deleted, channel:)

    login_as(user, scope: :user)

    expect(Program.count).to eq(1)

    delete "/db/programs/#{program.id}"
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(Program.count).to eq(1)
  end

  it "編集者ユーザーがログインしているとき、アクセス拒否されること" do
    user = create(:registered_user, :with_editor_role)
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = create(:program, :not_deleted, channel:)

    login_as(user, scope: :user)

    expect(Program.count).to eq(1)

    delete "/db/programs/#{program.id}"
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(Program.count).to eq(1)
  end

  it "管理者ユーザーがログインしているとき、プログラムをソフト削除できること" do
    user = create(:registered_user, :with_admin_role)
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = create(:program, :not_deleted, channel:)

    login_as(user, scope: :user)

    expect(Program.count).to eq(1)

    delete "/db/programs/#{program.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")

    expect(Program.count).to eq(0)
  end
end
