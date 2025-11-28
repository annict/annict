# typed: false
# frozen_string_literal: true

RSpec.describe "GET /", type: :request do
  it "ログインしているとき、アクティビティが存在しないとき、アクティビティが無い旨を表示すること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    get "/"

    expect(response.status).to eq(200)
    expect(response.body).to include("アクティビティはありません")
  end

  it "ログインしているとき、アクティビティが存在するとき、アクティビティを表示すること" do
    user = FactoryBot.create(:registered_user)
    episode_record = FactoryBot.create(:episode_record, user: user, body: "楽しかった")
    activity_group = FactoryBot.create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true)
    FactoryBot.create(:activity, user: user, itemable: episode_record, activity_group: activity_group)
    login_as(user, scope: :user)

    get "/"

    expect(response.status).to eq(200)
    expect(response.body).to include(user.profile.name)
    expect(response.body).to include("が記録しました")
    expect(response.body).to include(episode_record.episode.title)
    expect(response.body).to include("楽しかった")
  end

  it "ログインしているとき、フォローしているユーザーのアクティビティを表示すること" do
    user = FactoryBot.create(:registered_user)
    following_user = FactoryBot.create(:registered_user)
    FactoryBot.create(:follow, user: user, following: following_user)
    episode_record = FactoryBot.create(:episode_record, user: following_user, body: "面白かった")
    activity_group = FactoryBot.create(:activity_group, user: following_user, itemable_type: "EpisodeRecord", single: true)
    FactoryBot.create(:activity, user: following_user, itemable: episode_record, activity_group: activity_group)
    login_as(user, scope: :user)

    get "/"

    expect(response.status).to eq(200)
    expect(response.body).to include(following_user.profile.name)
    expect(response.body).to include("が記録しました")
    expect(response.body).to include(episode_record.episode.title)
    expect(response.body).to include("面白かった")
  end

  it "ログインしているとき、ページネーションが機能すること" do
    user = FactoryBot.create(:registered_user)
    following_user = FactoryBot.create(:registered_user)
    FactoryBot.create(:follow, user: user, following: following_user)

    31.times do |i|
      episode_record = FactoryBot.create(:episode_record, user: following_user)
      activity_group = FactoryBot.create(:activity_group, user: following_user, itemable_type: "EpisodeRecord", single: true)
      FactoryBot.create(:activity, user: following_user, itemable: episode_record, activity_group: activity_group)
    end

    login_as(user, scope: :user)

    get "/?page=2"

    expect(response.status).to eq(200)
    expect(response.body).to include(following_user.profile.name)
  end
end
