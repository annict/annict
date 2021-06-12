# frozen_string_literal: true

describe "GET /characters/:character_id", type: :request do
  let(:character) { create(:character) }

  it "アクセスできること" do
    get "/characters/#{character.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
  end
end
