# frozen_string_literal: true

describe "DELETE /v1/me/reviews/:id" do
  let(:user) { create(:user, :with_profile, :with_setting) }
  let(:application) { create(:oauth_application, owner: user) }
  let(:access_token) { create(:oauth_access_token, application: application) }
  let(:anime) { create(:anime, :with_current_season) }
  let!(:record) { create(:record, anime: anime, user: user) }
  let!(:anime_record) { create(:anime_record, anime: anime, user: user) }

  it "responses 204" do
    delete api("/v1/me/reviews/#{anime_record.id}", access_token: access_token.token)
    expect(response.status).to eq(204)
  end

  it "deletes a record" do
    expect(access_token.owner.anime_records.count).to eq(1)

    delete api("/v1/me/reviews/#{anime_record.id}", access_token: access_token.token)

    expect(access_token.owner.anime_records.count).to eq(0)
  end
end
