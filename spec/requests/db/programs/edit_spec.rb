# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/programs/:id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = FactoryBot.create(:program, channel:)

    get "/db/programs/#{program.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限がないユーザーの場合、アクセスできないこと" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    user = FactoryBot.create(:registered_user)
    program = FactoryBot.create(:program, channel:)

    login_as(user, scope: :user)
    get "/db/programs/#{program.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限があるユーザーの場合、番組編集フォームが表示されること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    program = FactoryBot.create(:program, channel:)

    login_as(user, scope: :user)
    get "/db/programs/#{program.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(program.channel.name)
  end

  it "管理者権限があるユーザーの場合、番組編集フォームが表示されること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    user = FactoryBot.create(:registered_user, :with_admin_role)
    program = FactoryBot.create(:program, channel:)

    login_as(user, scope: :user)
    get "/db/programs/#{program.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(program.channel.name)
  end

  it "存在しない番組IDの場合、404エラーが発生すること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)

    login_as(user, scope: :user)

    expect do
      get "/db/programs/non-existent-id/edit"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済み番組の場合、404エラーが発生すること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    program = FactoryBot.create(:program, channel:, deleted_at: Time.current)

    login_as(user, scope: :user)

    expect do
      get "/db/programs/#{program.id}/edit"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end
end
