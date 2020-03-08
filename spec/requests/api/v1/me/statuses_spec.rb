# frozen_string_literal: true

describe "API::V1::Me::Statuses" do
  let(:access_token) { create(:oauth_access_token) }
  let(:work) { create(:work, :with_current_season) }
  let!(:tip) { create(:status_tip) }

  describe "POST /v1/me/statuses", debug: true do
    before do
      data = {
        work_id: work.id,
        kind: "watching",
        access_token: access_token.token
      }
      post api("/v1/me/statuses", data)
    end

    it "responses 204" do
      expect(response.status).to eq(204)
    end

    it "saves status info" do
      expect(access_token.owner.statuses.count).to eq(1)
      expect(access_token.owner.statuses.first.kind).to eq("watching")
    end
  end
end
