# frozen_string_literal: true

describe "Api::V1::Me::Records" do
  let(:user) { create(:user, :with_profile, :with_setting) }
  let(:application) { create(:oauth_application, owner: user) }
  let(:access_token) { create(:oauth_access_token, application: application) }
  let(:work) { create(:work, :with_current_season) }
  let(:episode) { create(:episode, work: work) }
  let!(:tip) { create(:record_tip) }

  describe "POST /v1/me/records" do
    before do
      data = {
        episode_id: episode.id,
        comment: "Hello World",
        access_token: access_token.token
      }
      post api("/v1/me/records", data)
    end

    it "200が返ること" do
      expect(response.status).to eq(200)
    end

    it "記録されること" do
      expect(access_token.owner.checkins.count).to eq(1)
      expect(access_token.owner.checkins.first.comment).to eq("Hello World")
    end
  end

  describe "PATCH /v1/me/records/:id" do
    let(:record) { create(:checkin, work: work, episode: episode, user: user) }
    let(:uniq_comment) { SecureRandom.uuid }

    before do
      data = {
        comment: uniq_comment,
        access_token: access_token.token
      }
      patch api("/v1/me/records/#{record.id}", data)
    end

    it "200が返ること" do
      expect(response.status).to eq(200)
    end

    it "記録が更新されること" do
      expect(access_token.owner.checkins.count).to eq(1)
      expect(access_token.owner.checkins.first.comment).to eq(uniq_comment)
    end
  end

  describe "DELETE /v1/me/records/:id" do
    let!(:record) { create(:checkin, work: work, episode: episode, user: user) }

    it "204が返ること" do
      delete api("/v1/me/records/#{record.id}", access_token: access_token.token)
      expect(response.status).to eq(204)
    end

    it "記録が削除されること" do
      expect(access_token.owner.checkins.count).to eq(1)

      delete api("/v1/me/records/#{record.id}", access_token: access_token.token)

      expect(access_token.owner.checkins.count).to eq(0)
    end
  end
end
