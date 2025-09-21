# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/programs/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = create(:program, :unpublished, channel:)

    post "/db/programs/#{program.id}/publishing"
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(program.published?).to eq(false)
  end

  it "編集者でないユーザーがログインしているとき、アクセスが拒否されること" do
    user = create(:registered_user)
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = create(:program, :unpublished, channel:)

    login_as(user, scope: :user)

    post "/db/programs/#{program.id}/publishing"
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(program.published?).to eq(false)
  end

  it "編集者がログインしているとき、プログラムを公開できること" do
    user = create(:registered_user, :with_editor_role)
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = create(:program, :unpublished, channel:)

    login_as(user, scope: :user)

    expect(program.published?).to eq(false)

    post "/db/programs/#{program.id}/publishing"
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(program.published?).to eq(true)
  end

  it "編集者がログインしているとき、存在しないプログラムIDを指定すると404エラーが発生すること" do
    user = create(:registered_user, :with_editor_role)

    login_as(user, scope: :user)

    post "/db/programs/nonexistent-id/publishing"

    expect(response.status).to eq(404)
  end

  it "編集者がログインしているとき、既に公開済みのプログラムを指定すると404エラーが発生すること" do
    user = create(:registered_user, :with_editor_role)
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = create(:program, :published, channel:)

    login_as(user, scope: :user)

    post "/db/programs/#{program.id}/publishing"

    expect(response.status).to eq(404)
  end
end
