# typed: false
# frozen_string_literal: true

RSpec.describe "GET /fragment/@:username/records", type: :request do
  it "ユーザーが存在するとき、記録一覧を表示すること" do
    user = FactoryBot.create(:registered_user, username: "testuser")
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work: work)
    record_with_episode = FactoryBot.create(:record, :with_episode_record, user: user, work: work, episode: episode)
    record_with_episode.episode_record.update!(body: "面白かった")
    record_with_work = FactoryBot.create(:record, :with_work_record, user: user, work: work)
    record_with_work.work_record.update!(body: "素晴らしい作品でした")

    get "/fragment/@testuser/records"

    expect(response.status).to eq(200)
    expect(response.body).to include(episode.title)
    expect(response.body).to include("面白かった")
    expect(response.body).to include("素晴らしい作品でした")
  end

  it "ユーザーが存在しないとき、404エラーを返すこと" do
    get "/fragment/@nonexistentuser/records"
    
    expect(response.status).to eq(404)
  end

  it "削除済みユーザーの記録にアクセスしたとき、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user, username: "deleteduser")
    user.destroy!

    get "/fragment/@deleteduser/records"
    
    expect(response.status).to eq(404)
  end

  it "記録が削除済みのとき、表示されないこと" do
    user = FactoryBot.create(:registered_user, username: "testuser")
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work: work)
    record1 = FactoryBot.create(:record, :with_episode_record, user: user, work: work, episode: episode)
    record1.episode_record.update!(body: "表示される記録")
    record2 = FactoryBot.create(:record, :with_episode_record, user: user, work: work, episode: episode)
    record2.episode_record.update!(body: "削除された記録")
    record2.destroy!

    get "/fragment/@testuser/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("表示される記録")
    expect(response.body).not_to include("削除された記録")
  end

  it "他のユーザーの記録が混ざらないこと" do
    user1 = FactoryBot.create(:registered_user, username: "user1")
    user2 = FactoryBot.create(:registered_user, username: "user2")
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work: work)
    record1 = FactoryBot.create(:record, :with_episode_record, user: user1, work: work, episode: episode)
    record1.episode_record.update!(body: "ユーザー1の記録")
    record2 = FactoryBot.create(:record, :with_episode_record, user: user2, work: work, episode: episode)
    record2.episode_record.update!(body: "ユーザー2の記録")

    get "/fragment/@user1/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("ユーザー1の記録")
    expect(response.body).not_to include("ユーザー2の記録")
  end

  it "記録がないとき、空の一覧を表示すること" do
    FactoryBot.create(:registered_user, username: "emptyuser")

    get "/fragment/@emptyuser/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("記録はありません")
  end
end
