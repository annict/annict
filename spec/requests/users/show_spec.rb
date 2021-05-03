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
      let!(:record_1) { create(:record, :with_episode_record, user: user) }
      let!(:record_2) { create(:record, :with_anime_record, user: user) }
      let!(:status_activity) { create(:activity, :with_activity_group, user: user, itemable: status) }
      let!(:episode_record_activity) { create(:activity, :with_activity_group, user: user, itemable: record_1.episode_record) }
      let!(:anime_record_activity) { create(:activity, :with_activity_group, user: user, itemable: record_2.anime_record) }

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
      let!(:record_1) { create(:record, :with_episode_record, user: user) }
      let!(:record_2) { create(:record, :with_anime_record, user: user) }
      let!(:status_activity) { create(:activity, :with_activity_group, user: user, itemable: status) }
      let!(:episode_record_activity) { create(:activity, :with_activity_group, user: user, itemable: record_1.episode_record) }
      let!(:anime_record_activity) { create(:activity, :with_activity_group, user: user, itemable: record_2.anime_record) }

      it "アクティビティが表示されること" do
        get "/@#{user.username}"

        expect(response.status).to eq(200)
        expect(response.body).to include("がステータスを変更しました")
        expect(response.body).to include("が記録しました")
      end
    end
  end
end
