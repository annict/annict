# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/programs/:id/publishing", type: :request do
  it "ログインしていないとき、アクセスできず番組の公開状態が変更されないこと" do
    channel_group = create(:channel_group)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = create(:program, :published, channel:)

    delete "/db/programs/#{program.id}/publishing"
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(program.published?).to eq(true)
  end

  it "編集者ではないユーザーがログインしているとき、アクセスできず番組の公開状態が変更されないこと" do
    user = create(:registered_user)
    channel_group = create(:channel_group)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = create(:program, :published, channel:)
    login_as(user, scope: :user)

    delete "/db/programs/#{program.id}/publishing"
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(program.published?).to eq(true)
  end

  it "編集者がログインしているとき、番組を非公開にできること" do
    user = create(:registered_user, :with_editor_role)
    channel_group = create(:channel_group)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = create(:program, :published, channel:)
    login_as(user, scope: :user)

    expect(program.published?).to eq(true)

    delete "/db/programs/#{program.id}/publishing"
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(program.published?).to eq(false)
  end
end
