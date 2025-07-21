# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/programs", type: :request do
  it "ユーザーがログインしていないとき、番組一覧が表示されること" do
    program = create(:program)

    get "/db/works/#{program.work_id}/programs"

    expect(response.status).to eq(200)
    expect(response.body).to include(program.channel.name)
  end

  it "ユーザーがログインしているとき、番組一覧が表示されること" do
    user = create(:registered_user)
    program = create(:program)
    login_as(user, scope: :user)

    get "/db/works/#{program.work_id}/programs"

    expect(response.status).to eq(200)
    expect(response.body).to include(program.channel.name)
  end

  it "削除されたプログラムは表示されないこと" do
    work = create(:work)
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel1 = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    channel2 = Channel.create!(channel_group:, name: "フジテレビ", sort_number: 2)
    program = create(:program, work:, channel: channel1)
    deleted_program = create(:program, :deleted, work:, channel: channel2)

    get "/db/works/#{work.id}/programs"

    expect(response.status).to eq(200)
    expect(response.body).to include(program.channel.name)
    expect(response.body).not_to include(deleted_program.channel.name)
  end

  it "存在しない作品のプログラム一覧にアクセスしたとき、ActiveRecord::RecordNotFoundエラーが発生すること" do
    non_existent_work_id = "non-existent-id"

    expect {
      get "/db/works/#{non_existent_work_id}/programs"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除された作品のプログラム一覧にアクセスしたとき、ActiveRecord::RecordNotFoundエラーが発生すること" do
    deleted_work = create(:work, :deleted)

    expect {
      get "/db/works/#{deleted_work.id}/programs"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
