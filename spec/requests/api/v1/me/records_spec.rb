# frozen_string_literal: true

describe "API::V1::Me::Records" do
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

    it "responses 200" do
      expect(response.status).to eq(200)
    end

    it "creates a record" do
      expect(access_token.owner.episode_records.count).to eq(1)
      expect(access_token.owner.episode_records.first.body).to eq("Hello World")
    end
  end

  describe "PATCH /v1/me/records/:id" do
    let(:record) { create(:episode_record, work: work, episode: episode, user: user) }
    let(:uniq_comment) { SecureRandom.uuid }

    before do
      data = {
        comment: uniq_comment,
        access_token: access_token.token
      }
      patch api("/v1/me/records/#{record.id}", data)
    end

    it "responses 200" do
      expect(response.status).to eq(200)
    end

    it "updates a record" do
      expect(access_token.owner.episode_records.count).to eq(1)
      expect(access_token.owner.episode_records.first.body).to eq(uniq_comment)
    end
  end

  describe "DELETE /v1/me/records/:id" do
    let!(:record) { create(:episode_record, work: work, episode: episode, user: user) }

    it "responses 204" do
      delete api("/v1/me/records/#{record.id}", access_token: access_token.token)
      expect(response.status).to eq(204)
    end

    it "deletes a record" do
      expect(access_token.owner.episode_records.count).to eq(1)

      delete api("/v1/me/records/#{record.id}", access_token: access_token.token)

      expect(access_token.owner.episode_records.count).to eq(0)
    end
  end
end
