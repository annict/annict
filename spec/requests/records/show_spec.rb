# typed: false
# frozen_string_literal: true

RSpec.describe "GET /@:username/records/:record_id", type: :request do
  it "ログインしてアニメへの記録を参照するとき、記録が表示されること" do
    user = create(:registered_user)
    work = create(:work)
    record = create(:record, user:, work:)
    create(:work_record, user:, record:, work:, body: "最高")

    login_as(user, scope: :user)
    get "/@#{user.username}/records/#{record.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(user.profile.name)
    expect(response.body).to include(work.title)
    expect(response.body).to include("最高")
  end

  it "ログインしてエピソードへの記録を参照するとき、記録が表示されること" do
    user = create(:registered_user)
    work = create(:work)
    episode = create(:episode, work:)
    record = create(:record, user:, work:)
    create(:episode_record, record:, work:, episode:, user:, body: "楽しかった")

    login_as(user, scope: :user)
    get "/@#{user.username}/records/#{record.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(user.profile.name)
    expect(response.body).to include(work.title)
    expect(response.body).to include(episode.number)
    expect(response.body).to include("楽しかった")
  end

  it "ログインしていないときアニメへの記録を参照すると、記録が表示されること" do
    user = create(:registered_user)
    work = create(:work)
    record = create(:record, user:, work:)
    create(:work_record, user:, record:, work:, body: "最高")

    get "/@#{user.username}/records/#{record.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(user.profile.name)
    expect(response.body).to include(work.title)
    expect(response.body).to include("最高")
  end

  it "ログインしていないときエピソードへの記録を参照すると、記録が表示されること" do
    user = create(:registered_user)
    work = create(:work)
    episode = create(:episode, work:)
    record = create(:record, user:, work:)
    create(:episode_record, record:, work:, episode:, user:, body: "楽しかった")

    get "/@#{user.username}/records/#{record.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(user.profile.name)
    expect(response.body).to include(work.title)
    expect(response.body).to include(episode.number)
    expect(response.body).to include("楽しかった")
  end

  it "存在しない記録を参照すると、例外が発生すること" do
    user = create(:registered_user)

    expect {
      get "/@#{user.username}/records/non-existent-id"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
