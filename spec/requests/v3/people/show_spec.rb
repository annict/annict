# frozen_string_literal: true

describe "GET /people/:person_id", type: :request do
  let(:person) { create(:person) }

  it "アクセスできること" do
    get "/people/#{person.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(person.name)
  end
end
