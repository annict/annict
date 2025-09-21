# typed: false
# frozen_string_literal: true

RSpec.describe "GET /fragment/activity_groups/:activity_group_id/items", type: :request do
  it "ログインしているとき、アクティビティに紐付く記録を表示すること" do
    user = create(:registered_user)
    episode_record = create(:episode_record, user: user, body: "楽しかった")
    activity_group = create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true)
    create(:activity, user: user, itemable: episode_record, activity_group: activity_group)

    login_as(user, scope: :user)
    get "/fragment/activity_groups/#{activity_group.id}/items"

    expect(response.status).to eq(200)
    expect(response.body).to include(episode_record.episode.title)
    expect(response.body).to include("楽しかった")
  end

  it "ログインしていないとき、アクティビティが紐付いていればアクセスできること" do
    user = create(:registered_user)
    episode_record = create(:episode_record, user: user, body: "面白かった")
    activity_group = create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true)
    create(:activity, user: user, itemable: episode_record, activity_group: activity_group)

    get "/fragment/activity_groups/#{activity_group.id}/items"

    expect(response.status).to eq(200)
    expect(response.body).to include(episode_record.episode.title)
    expect(response.body).to include("面白かった")
  end

  it "存在しないactivity_groupを指定したとき、404エラーが返されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect {
      get "/fragment/activity_groups/non-existent-id/items"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
