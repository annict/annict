# typed: false
# frozen_string_literal: true

describe "GET /", type: :request do
  context "ログインしているとき" do
    let!(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    context "アクティビティが存在しないとき" do
      it "アクティビティが無い旨を表示すること" do
        get "/"

        expect(response.status).to eq(200)
        expect(response.body).to include("アクティビティはありません")
      end
    end

    context "アクティビティが存在するとき" do
      let!(:episode_record) { create(:episode_record, user: user, body: "楽しかった") }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true) }
      let!(:activity) { create(:activity, user: user, itemable: episode_record, activity_group: activity_group) }

      it "アクティビティを表示すること" do
        get "/"

        expect(response.status).to eq(200)
        expect(response.body).to include(user.profile.name)
        expect(response.body).to include("が記録しました")
        expect(response.body).to include(episode_record.episode.title)
        expect(response.body).to include("楽しかった")
      end
    end
  end
end
