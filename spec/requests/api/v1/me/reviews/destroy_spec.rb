# frozen_string_literal: true

describe "DELETE /v1/me/reviews/:id" do
  let(:user) { create(:user, :with_profile, :with_setting) }
  let(:application) { create(:oauth_application, owner: user) }
  let(:access_token) { create(:oauth_access_token, application: application) }
  let(:work) { create(:work, :with_current_season) }
  let!(:work_record) { create(:work_record) }
  let!(:record) { create(:record, :for_work, user: user, work: work, recordable: work_record) }

  it "responses 204" do
    delete api("/v1/me/reviews/#{work_record.id}", access_token: access_token.token)
    expect(response.status).to eq(204)
  end

  it "deletes a record" do
    expect(access_token.owner.records.count).to eq(1)

    delete api("/v1/me/reviews/#{work_record.id}", access_token: access_token.token)

    expect(access_token.owner.records.count).to eq(0)
  end
end
