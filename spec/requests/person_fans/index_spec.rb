# typed: false
# frozen_string_literal: true

describe "GET /people/:person_id/fans", type: :request do
  let(:person) { create(:person) }
  let!(:user) { create(:registered_user) }
  let!(:person_favorite) { create(:person_favorite, user: user, person: person) }

  it "アクセスできること" do
    get "/people/#{person.id}/fans"

    expect(response.status).to eq(200)
    expect(response.body).to include(person.name)
    expect(response.body).to include(user.profile.name)
  end
end
