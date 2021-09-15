# frozen_string_literal: true

describe "PATCH /v1/me/reviews/:id" do
  let(:user) { create(:user, :with_profile, :with_setting) }
  let(:application) { create(:oauth_application, owner: user) }
  let(:access_token) { create(:oauth_access_token, application: application) }
  let(:work) { create(:work, :with_current_season) }
  let(:work_record) { create(:work_record) }
  let!(:record) { create(:record, :on_work, user: user, work: work, recordable: work_record) }
  let(:uniq_title) { SecureRandom.uuid }
  let(:uniq_body) { SecureRandom.uuid }

  before do
    data = {
      title: uniq_title,
      body: uniq_body,
      access_token: access_token.token
    }
    patch api("/v1/me/reviews/#{work_record.id}", data)
  end

  it "responses 200" do
    expect(response.status).to eq(200)
  end

  it "updates an work record" do
    expected_body = "#{uniq_title}\n\n#{uniq_body}"
    expect(access_token.owner.records.count).to eq(1)
    expect(access_token.owner.records.first.body).to eq(expected_body)
    expect(json["body"]).to eq(expected_body)
  end
end
