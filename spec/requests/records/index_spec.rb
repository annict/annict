# frozen_string_literal: true

describe "GET /@:username/records", type: :request do
  context "when user signs in" do
    let!(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    context "when records do not exist" do
      it "displays no exists message" do
        get "/@#{user.username}/records"

        expect(response.status).to eq(200)
        expect(response.body).to include("記録はありません")
      end
    end

    context "when records exist" do
      let!(:episode_record) { create(:episode_record) }
      let!(:record_1) { create(:record, :for_episode, user: user, recordable: episode_record, body: "楽しかった") }
      let!(:work_record) { create(:work_record) }
      let!(:record_2) { create(:record, user: user, recordable: work_record, body: "最高") }

      it "displays records" do
        get "/@#{user.username}/records"

        expect(response.status).to eq(200)
        expect(response.body).to include(user.profile.name)
        expect(response.body).to include("楽しかった")
        expect(response.body).to include("最高")
      end
    end
  end

  context "when user does not sign in" do
    let!(:user) { create(:registered_user) }

    context "when records do not exist" do
      it "displays no exists message" do
        get "/@#{user.username}/records"

        expect(response.status).to eq(200)
        expect(response.body).to include("記録はありません")
      end
    end

    context "when records exist" do
      let!(:episode_record) { create(:episode_record) }
      let!(:record_1) { create(:record, :for_episode, user: user, recordable: episode_record, body: "楽しかった") }
      let!(:work_record) { create(:work_record) }
      let!(:record_2) { create(:record, :for_work, user: user, recordable: work_record, body: "最高") }

      it "displays records" do
        get "/@#{user.username}/records"

        expect(response.status).to eq(200)
        expect(response.body).to include(user.profile.name)
        expect(response.body).to include("楽しかった")
        expect(response.body).to include("最高")
      end
    end
  end

  context "when the month parameter is attached" do
    let!(:user) { create(:registered_user) }
    let!(:record_1_work_record) { create(:work_record) }
    let!(:record_1) { create(:record, :for_work, user: user, recordable: record_1_work_record, body: "最高", watched_at: Time.zone.parse("2020-04-01")) }
    let!(:record_2_work_record) { create(:work_record) }
    let!(:record_2) { create(:record, :for_work, user: user, recordable: record_2_work_record, body: "すごく良かった", created_at: Time.zone.parse("2020-05-01")) }

    it "displays records" do
      get "/@#{user.username}/records?year=2020&month=5"

      expect(response.status).to eq(200)
      expect(response.body).to include(user.profile.name)
      expect(response.body).to include("すごく良かった")
      expect(response.body).to_not include("最高")
    end
  end
end
