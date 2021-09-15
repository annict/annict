# frozen_string_literal: true

describe "GET /fragment/activity_groups/:activity_group_id/items", type: :request do
  context "ログインしているとき" do
    let!(:user) { create(:registered_user) }
    let!(:record) { create(:record, :on_episode, user: user, body: "楽しかった") }
    let!(:activity_group) { create(:activity_group, user: user, itemable_type: "Record", single: true) }
    let!(:activity) { create(:activity, user: user, itemable: record, activity_group: activity_group) }

    before do
      login_as(user, scope: :user)
    end

    it "アクティビティに紐付く記録を表示すること" do
      get "/fragment/activity_groups/#{activity_group.id}/items"

      expect(response.status).to eq(200)
      expect(response.body).to include(record.episode.title)
      expect(response.body).to include("楽しかった")
    end
  end
end
