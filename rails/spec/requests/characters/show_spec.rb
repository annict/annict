# typed: false
# frozen_string_literal: true

RSpec.describe "GET /characters/:character_id", type: :request do
  it "アクセスできること" do
    character = FactoryBot.create(:character)

    get "/characters/#{character.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
  end
end
