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
      let!(:record_1) { create(:record, user: user) }
      let!(:episode_record) { create(:episode_record, record: record_1, user: user, body: "楽しかった") }
      let!(:record_2) { create(:record, user: user) }
      let!(:work_record) { create(:work_record, user: user, record: record_2, body: "最高") }

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
      let!(:record_1) { create(:record, user: user) }
      let!(:episode_record) { create(:episode_record, record: record_1, user: user, body: "楽しかった") }
      let!(:record_2) { create(:record, user: user) }
      let!(:work_record) { create(:work_record, user: user, record: record_2, body: "最高") }

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
    let!(:record_1) { create(:record, user: user, watched_at: Time.zone.parse("2020-04-01")) }
    let!(:record_1_work_record) { create(:work_record, user: user, record: record_1, body: "最高") }
    let!(:record_2) { create(:record, user: user, watched_at: Time.zone.parse("2020-05-01")) }
    let!(:record_2_work_record) { create(:work_record, user: user, record: record_2, body: "すごく良かった") }

    it "displays records" do
      get "/@#{user.username}/records?year=2020&month=5"

      expect(response.status).to eq(200)
      expect(response.body).to include(user.profile.name)
      expect(response.body).to include("すごく良かった")
      expect(response.body).to_not include("最高")
    end
  end
end
