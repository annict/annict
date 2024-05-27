# typed: false
# frozen_string_literal: true

describe "GET /@:username", type: :request do
  let(:user) { create(:registered_user) }

  context "ログインしているとき" do
    before do
      login_as(user, scope: :user)
    end

    context "アクティビティが存在しないとき" do
      it "EmptyComponent が表示されること" do
        get "/@#{user.username}"

        expect(response.status).to eq(200)
        expect(response.body).to include("アクティビティはありません")
      end
    end

    context "アクティビティが存在するとき" do
      let!(:status) { create(:status, user: user) }
      let!(:status_activity_group) { create(:activity_group, user: user, itemable_type: "Status", single: false) }
      let!(:status_activity) { create(:activity, user: user, itemable: status, activity_group: status_activity_group) }

      let!(:record_1) { create(:record, :with_episode_record, user: user) }
      let!(:episode_record_activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true) }
      let!(:episode_record_activity) { create(:activity, user: user, itemable: record_1.episode_record, activity_group: episode_record_activity_group) }

      let!(:record_2) { create(:record, :with_work_record, user: user) }
      let!(:work_record_activity_group) { create(:activity_group, user: user, itemable_type: "WorkRecord", single: true) }
      let!(:work_record_activity) { create(:activity, user: user, itemable: record_2.work_record, activity_group: work_record_activity_group) }

      it "アクティビティが表示されること" do
        get "/@#{user.username}"

        expect(response.status).to eq(200)
        expect(response.body).to include("がステータスを変更しました")
        expect(response.body).to include("が記録しました")
      end
    end
  end

  context "ログインしていないとき" do
    context "アクティビティが存在しないとき" do
      it "EmptyComponent が表示されること" do
        get "/@#{user.username}"

        expect(response.status).to eq(200)
        expect(response.body).to include("アクティビティはありません")
      end
    end

    context "アクティビティが存在するとき" do
      let!(:status) { create(:status, user: user) }
      let!(:status_activity_group) { create(:activity_group, user: user, itemable_type: "Status", single: false) }
      let!(:status_activity) { create(:activity, user: user, itemable: status, activity_group: status_activity_group) }

      let!(:record_1) { create(:record, :with_episode_record, user: user) }
      let!(:episode_record_activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true) }
      let!(:episode_record_activity) { create(:activity, user: user, itemable: record_1.episode_record, activity_group: episode_record_activity_group) }

      let!(:record_2) { create(:record, :with_work_record, user: user) }
      let!(:work_record_activity_group) { create(:activity_group, user: user, itemable_type: "WorkRecord", single: true) }
      let!(:work_record_activity) { create(:activity, user: user, itemable: record_2.work_record, activity_group: work_record_activity_group) }

      it "アクティビティが表示されること" do
        get "/@#{user.username}"

        expect(response.status).to eq(200)
        expect(response.body).to include("がステータスを変更しました")
        expect(response.body).to include("が記録しました")
      end
    end
  end
end
