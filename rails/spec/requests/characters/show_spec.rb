# typed: false
# frozen_string_literal: true

RSpec.describe "GET /characters/:character_id", type: :request do
  it "アクセスできること" do
    character = FactoryBot.create(:character)

    get "/characters/#{character.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
  end

  it "人物が非公開のキャストの声優は表示されないこと" do
    character = FactoryBot.create(:character)
    published_person = FactoryBot.create(:person, name: "公開中の声優")
    unpublished_person = FactoryBot.create(:person, :unpublished, name: "非公開の声優")
    FactoryBot.create(:cast, character:, person: published_person, name: published_person.name)
    FactoryBot.create(:cast, character:, person: unpublished_person, name: unpublished_person.name)

    get "/characters/#{character.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include("公開中の声優")
    expect(response.body).not_to include("非公開の声優")
  end
end
