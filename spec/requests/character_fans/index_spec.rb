# typed: false
# frozen_string_literal: true

describe "GET /characters/:character_id/fans", type: :request do
  let(:character) { create(:character) }
  let!(:user) { create(:registered_user) }
  let!(:character_favorite) { create(:character_favorite, user: user, character: character) }

  it "アクセスできること" do
    get "/characters/#{character.id}/fans"

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
    expect(response.body).to include(user.profile.name)
  end
end
