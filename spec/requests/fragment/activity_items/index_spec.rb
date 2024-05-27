# typed: false
# frozen_string_literal: true

describe "GET /fragment/activity_groups/:activity_group_id/items", type: :request do
  context "ログインしているとき" do
    let!(:user) { create(:registered_user) }
    let!(:episode_record) { create(:episode_record, user: user, body: "楽しかった") }
    let!(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true) }
    let!(:activity) { create(:activity, user: user, itemable: episode_record, activity_group: activity_group) }

    before do
      login_as(user, scope: :user)
    end

    it "アクティビティに紐付く記録を表示すること" do
      get "/fragment/activity_groups/#{activity_group.id}/items"

      expect(response.status).to eq(200)
      expect(response.body).to include(episode_record.episode.title)
      expect(response.body).to include("楽しかった")
    end
  end
end
